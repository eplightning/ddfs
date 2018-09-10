package index

import (
	"os"
	"sync"

	"github.com/dgraph-io/badger"
	"github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"

	"git.eplight.org/eplightning/ddfs/pkg/monitor"
	"git.eplight.org/eplightning/ddfs/pkg/util"
)

type ShardManager struct {
	mon       monitor.Client
	shards    map[string]*Shard
	mutex     sync.RWMutex
	cache     *lru.ARCCache
	db        *badger.DB
	cacheSize int
	dataPath  string
}

func NewShardManager(mon monitor.Client, cacheSize int, dataPath string) *ShardManager {
	return &ShardManager{
		mon:       mon,
		shards:    make(map[string]*Shard),
		cacheSize: cacheSize,
		dataPath:  dataPath,
	}
}

func (s *ShardManager) Init() error {
	if err := os.MkdirAll(s.dataPath, 0755); err != nil {
		return err
	}

	opt := badger.Options{
		Dir:      s.dataPath,
		ValueDir: s.dataPath,
	}

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

	return nil
}

func (s *ShardManager) Start(ctl *util.SubsystemControl) {
	ctl.WaitGroup.Add(1)
	go func() {
		defer ctl.WaitGroup.Done()
		defer s.db.Close()
		defer func() {
			s.mutex.Lock()
			defer s.mutex.Unlock()
			s.shards = make(map[string]*Shard)
		}()

		for {
			select {
			case <-ctl.Stop:
				return
			}
		}
	}()
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
	return true
}
