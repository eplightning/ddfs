package monitor

import (
	"context"
	"errors"
	"time"

	"git.eplight.org/eplightning/ddfs/pkg/api"

	"git.eplight.org/eplightning/ddfs/pkg/util"
	"github.com/gogo/protobuf/proto"
	"github.com/rs/zerolog/log"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/namespace"
)

type EtcdManager struct {
	servers []string
	prefix  string
	client  *clientv3.Client
	kv      clientv3.KV
	state   clusterState
}

type clusterState struct {
	volumes  map[string]clusterVolume
	startRev int64
}

type clusterVolume struct {
	data     *api.Volume
	revision int64
}

func NewEtcdManager(servers []string, prefix string) *EtcdManager {
	return &EtcdManager{
		servers: servers,
		prefix:  prefix,
	}
}

func (m *EtcdManager) Init() error {
	cl, err := clientv3.New(clientv3.Config{
		Endpoints:   m.servers,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}

	m.client = cl
	m.client.KV = namespace.NewKV(m.client.KV, m.prefix)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := m.client.Get(ctx, "init")
	if err != nil {
		return err
	}

	if len(resp.Kvs) == 0 {
		log.Info().Msg("Initalizing new cluster")
		err = m.initMonitor(ctx)
		if err != nil {
			return err
		}
	}

	state, err := m.loadMonitor(ctx)
	if err != nil {
		return err
	}

	m.state = *state
	return nil
}

func (m *EtcdManager) Start(ctl *util.SubsystemControl) {
	ctl.WaitGroup.Add(1)
	go func() {
		defer ctl.WaitGroup.Done()
		defer m.client.Close()

		for {
			select {
			case <-ctl.Stop:
				return
			}
		}
	}()
}

func (m *EtcdManager) ClientSettings(ctx context.Context) (*api.ClientSettings, int64, error) {
	output := &api.ClientSettings{}
	rev, err := m.fetchAndUnmarshal(ctx, "settings/client", output)
	if err != nil {
		return nil, 0, err
	}

	return output, rev, nil
}

func (m *EtcdManager) fetchAndUnmarshal(ctx context.Context, key string, msg proto.Message) (int64, error) {
	resp, err := m.client.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if len(resp.Kvs) == 0 {
		return 0, errors.New("no client settings key")
	}

	err = proto.Unmarshal(resp.Kvs[0].Value, msg)
	if err != nil {
		return 0, err
	}

	return resp.Kvs[0].ModRevision, nil
}

func (m *EtcdManager) initMonitor(ctx context.Context) error {
	cs := &api.ClientSettings{
		MinFillSize: 1024,
		HashAlgo:    api.HashAlgorithm_SHA256,
		ChunkAlgo: &api.ClientSettings_Rabin{
			Rabin: &api.RabinChunkAlgorithm{
				MaxSize: 4 * 1024 * 1024,
				MinSize: 100 * 1024,
				Poly:    12313278162312893,
			},
		},
	}

	ss := &api.ServerSettings{
		ShardEntriesPerSlice: 1024,
		ShardSize:            256 * 1024 * 1024 * 1024,
	}

	blocks := &api.NodeReplicaSets{
		Sets: []*api.NodeReplicaSet{
			&api.NodeReplicaSet{
				Nodes: []*api.Node{
					&api.Node{
						Address: "localhost:7301",
						Name:    "first",
						Primary: true,
					},
				},
				Slot: 0,
			},
		},
	}

	indices := &api.NodeReplicaSets{
		Sets: []*api.NodeReplicaSet{
			&api.NodeReplicaSet{
				Nodes: []*api.Node{
					&api.Node{
						Address: "localhost:7302",
						Name:    "first",
						Primary: true,
					},
				},
				Slot: 0,
			},
		},
	}

	volume := &api.Volume{
		Name:            "first",
		RequestedShards: 1,
		Shards:          0,
		State:           api.VolumeState_NEW,
	}

	data := make(map[string]interface{})
	data["settings/client"] = cs
	data["settings/server"] = ss
	data["nodes/block"] = blocks
	data["nodes/index"] = indices
	data["volumes/first"] = volume

	for k, v := range data {
		bytes, err := proto.Marshal(v.(proto.Message))
		if err != nil {
			return err
		}
		_, err = m.client.Put(ctx, k, string(bytes))
		if err != nil {
			return err
		}
	}

	_, err := m.client.Put(ctx, "init", "true")
	return err
}

func (m *EtcdManager) loadMonitor(ctx context.Context) (*clusterState, error) {
	state := &clusterState{
		volumes: make(map[string]clusterVolume),
	}

	resp, err := m.client.Get(ctx, "volumes/", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	state.startRev = resp.Header.Revision

	for _, key := range resp.Kvs {
		v := clusterVolume{data: &api.Volume{}, revision: key.ModRevision}
		err := proto.Unmarshal(key.Value, v.data)
		if err != nil {
			return nil, err
		}

		state.volumes[v.data.Name] = v
		log.Info().Msgf("Loaded volume %v", v.data.Name)
	}

	return state, nil
}
