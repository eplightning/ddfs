package index

import (
	"context"
	"os"
	"strconv"
	"sync"

	"git.eplight.org/eplightning/ddfs/pkg/api"

	"github.com/dgraph-io/badger"
	"github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"git.eplight.org/eplightning/ddfs/pkg/monitor"
	"git.eplight.org/eplightning/ddfs/pkg/util"
)

type ShardManager struct {
	mon             monitor.Client
	shards          map[string]*Shard
	mutex           sync.RWMutex
	cache           *lru.ARCCache
	db              *badger.DB
	cacheSize       int
	dataPath        string
	context         context.Context
	cancel          context.CancelFunc
	watch           chan *api.Volume
	serverSettings  *api.ServerSettings
	volumeRevision  int64
	indexRing       util.HashRing
	name            string
	entriesPerSlice int32
}

func NewShardManager(mon monitor.Client, cacheSize int, entriesPerSlice int32, dataPath string, name string) *ShardManager {
	context, cancel := context.WithCancel(context.Background())

	return &ShardManager{
		mon:             mon,
		shards:          make(map[string]*Shard),
		cacheSize:       cacheSize,
		dataPath:        dataPath,
		context:         context,
		cancel:          cancel,
		name:            name,
		entriesPerSlice: entriesPerSlice,
	}
}

func (s *ShardManager) Init() error {
	if err := os.MkdirAll(s.dataPath, 0755); err != nil {
		return err
	}

	opt := badger.DefaultOptions
	opt.Dir = s.dataPath
	opt.ValueDir = s.dataPath

	db, err := badger.Open(opt)
	if err != nil {
		return err
	}

	cache, err := lru.NewARC(s.cacheSize)
	if err != nil {
		return err
	}

	s.db = db
	s.cache = cache

	s.watch = make(chan *api.Volume, 50)

	ss, err := s.mon.GetServerSettings(context.TODO(), &api.ServerSettingsRequest{})
	if err != nil {
		return err
	}
	s.serverSettings = ss.Settings

	idxNodes, err := s.mon.GetIndexStores(context.TODO(), &api.GetIndexStoresRequest{})
	if err != nil {
		return err
	}
	s.indexRing = util.NewHashRing(idxNodes.Data)

	resp, err := s.mon.Get(context.TODO(), &api.GetVolumesRequest{})
	if err != nil {
		return err
	}
	if resp.Header.Error != 0 {
		return errors.New(resp.Header.ErrorMsg)
	}
	s.volumeRevision = resp.Header.Revision

	for _, x := range resp.Volumes {
		err = s.handleVolumeUpdate(x)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ShardManager) Start(ctl *util.SubsystemControl) {
	ctl.WaitGroup.Add(2)
	go func() {
		defer ctl.WaitGroup.Done()

		subcontext, subcancel := context.WithCancel(context.Background())
		defer subcancel()
		watch, err := s.mon.Watch(subcontext)
		if err != nil {
			ctl.Error(err)
			return
		}
		err = watch.Send(&api.WatchVolumesRequest{
			Header: &api.WatchRequestHeader{
				StartRevision: s.volumeRevision + 1,
				End:           false,
			},
		})
		if err != nil {
			ctl.Error(err)
			return
		}

		ctl.WaitGroup.Add(1)
		go func() {
			defer ctl.WaitGroup.Done()
			for {
				vols, err := watch.Recv()
				if err != nil {
					break
				}
				log.Info().Msg("Volume watch notification")
				for _, v := range vols.Volumes {
					select {
					case s.watch <- v:
					case <-s.context.Done():
					}
				}
			}
			close(s.watch)
		}()

		<-s.context.Done()
		watch.Send(&api.WatchVolumesRequest{
			Header: &api.WatchRequestHeader{
				StartRevision: s.volumeRevision,
				End:           true,
			},
		})

	}()
	go func() {
		defer ctl.WaitGroup.Done()
		defer s.db.Close()
		defer func() {
			s.mutex.Lock()
			defer s.mutex.Unlock()
			s.shards = make(map[string]*Shard)
		}()
		defer s.cancel()

		for {
			select {
			case vol, more := <-s.watch:
				if !more {
					ctl.Error(errors.New("volume watch stopped"))
					return
				}
				if err := s.handleVolumeUpdate(vol); err != nil {
					ctl.Error(err)
					return
				}
			case <-ctl.Stop:
				return
			}
		}
	}()
}

func (s *ShardManager) handleVolumeUpdate(vol *api.Volume) error {
	log.Info().Msgf("Handling volume update for %v", vol.Name)

	added := make([]int64, 0)
	removed := make([]int64, 0)

	shardsTodo := make(map[int64]bool)
	for i := 0; i < int(vol.RequestedShards); i++ {
		if s.myShard(vol.Name, int64(i)) {
			shardsTodo[int64(i)] = true
		}
	}
	for _, x := range vol.CurrentShards {
		if !s.myShard(vol.Name, x) {
			continue
		}
		_, exists := shardsTodo[x]
		name := s.shardName(vol.Name, x)

		if exists {
			delete(shardsTodo, x)
			alreadyLoaded, _ := s.loadShard(name)
			if !alreadyLoaded {
				log.Info().Msgf("Loading shard %v", x)
			}
		} else {
			log.Info().Msgf("Removing shard %v", x)
			s.removeShard(name)
			removed = append(removed, x)
			// TODO: garbage collection
		}
	}
	for x := range shardsTodo {
		log.Info().Msgf("Creating shard %v", x)
		name := s.shardName(vol.Name, x)
		s.createShard(name)
		added = append(added, x)
	}

	if len(added) > 0 || len(removed) > 0 {
		resp, err := s.mon.UpdateShards(context.TODO(), &api.UpdateShardsRequest{
			AddedShards:   added,
			RemovedShards: removed,
			VolumeName:    vol.Name,
		})
		if err != nil {
			return err
		}
		if resp.Header.Error != 0 {
			return errors.New(resp.Header.ErrorMsg)
		}
	}

	return nil
}

func (s *ShardManager) removeShard(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var err error
	x, exists := s.shards[name]
	if exists {
		err = x.Discard()
	}
	delete(s.shards, name)
	return err
}

func (s *ShardManager) loadShard(name string) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, ok := s.shards[name]
	if ok {
		return true, nil
	}
	shard, err := NewShard(ShardSettings{
		ID:        name,
		New:       false,
		Size:      s.serverSettings.ShardSize,
		SliceSize: s.entriesPerSlice,
	}, s.db, s.cache)
	if err != nil {
		return false, err
	}
	s.shards[name] = shard
	return false, nil
}

func (s *ShardManager) createShard(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	shard, err := NewShard(ShardSettings{
		ID:        name,
		New:       true,
		Size:      s.serverSettings.ShardSize,
		SliceSize: s.entriesPerSlice,
	}, s.db, s.cache)
	if err != nil {
		return err
	}
	s.shards[name] = shard
	return nil
}

func (s *ShardManager) Shard(name string) (*Shard, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	shard, ok := s.shards[name]
	if !ok {
		return nil, errors.New("shard does not exist")
	}
	return shard, nil
}

func (s *ShardManager) myShard(volume string, id int64) bool {
	node := s.indexRing.Shard(s.shardName(volume, id))
	if node == nil || node.Name != s.name {
		return false
	}
	return true
}

func (s *ShardManager) shardName(volume string, id int64) string {
	return volume + "-" + strconv.FormatInt(id, 10)
}
