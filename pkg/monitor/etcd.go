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
	"go.etcd.io/etcd/mvcc/mvccpb"
)

type EtcdManager struct {
	servers        []string
	prefix         string
	client         *clientv3.Client
	bootstrapFile  string
	bootstrapForce bool
}

func NewEtcdManager(servers []string, prefix string, bootstrapFile string, bootstrapForce bool) *EtcdManager {
	return &EtcdManager{
		servers:        servers,
		prefix:         prefix,
		bootstrapFile:  bootstrapFile,
		bootstrapForce: bootstrapForce,
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
	m.client.Watcher = namespace.NewWatcher(m.client.Watcher, m.prefix)
	m.client.Lease = namespace.NewLease(m.client.Lease, m.prefix)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := m.client.Get(ctx, "init")
	if err != nil {
		return err
	}

	if len(resp.Kvs) == 0 || m.bootstrapForce {
		log.Info().Msg("Initalizing new cluster")
		var bootstrapData BootstrapData
		if m.bootstrapFile == "" {
			bootstrapData = DefaultBootstrapData()
		} else {
			bootstrapData, err = LoadBootstrapFromFile(m.bootstrapFile)
			if err != nil {
				return err
			}
		}
		err = m.initMonitor(ctx, bootstrapData)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *EtcdManager) Start(ctl *util.SubsystemControl) {
	ctl.WaitGroup.Add(1)
	go func() {
		defer ctl.WaitGroup.Done()
		defer m.client.Close()
		<-ctl.Stop
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

func (m *EtcdManager) ServerSettings(ctx context.Context) (*api.ServerSettings, int64, error) {
	output := &api.ServerSettings{}
	rev, err := m.fetchAndUnmarshal(ctx, "settings/server", output)
	if err != nil {
		return nil, 0, err
	}

	return output, rev, nil
}

func (m *EtcdManager) BlockStores(ctx context.Context) (*api.NodeReplicaSets, int64, error) {
	output := &api.NodeReplicaSets{}
	rev, err := m.fetchAndUnmarshal(ctx, "nodes/block", output)
	if err != nil {
		return nil, 0, err
	}

	return output, rev, nil
}

func (m *EtcdManager) IndexStores(ctx context.Context) (*api.NodeReplicaSets, int64, error) {
	output := &api.NodeReplicaSets{}
	rev, err := m.fetchAndUnmarshal(ctx, "nodes/index", output)
	if err != nil {
		return nil, 0, err
	}

	return output, rev, nil
}

func (m *EtcdManager) Volumes(ctx context.Context) ([]*api.Volume, int64, error) {
	resp, err := m.client.Get(ctx, "volumes/", clientv3.WithPrefix())
	if err != nil {
		return nil, 0, err
	}

	volumes := make([]*api.Volume, len(resp.Kvs))
	var revision int64

	for i, key := range resp.Kvs {
		if key.ModRevision > revision {
			revision = key.ModRevision
		}

		volumes[i] = &api.Volume{}
		err := proto.Unmarshal(key.Value, volumes[i])
		if err != nil {
			return nil, 0, err
		}
	}

	return volumes, revision, nil
}

func (m *EtcdManager) WatchVolumes(ctx context.Context, start int64) chan []*api.Volume {
	output := make(chan []*api.Volume, 100)

	go func() {
		events := m.client.Watch(ctx, "volumes/", clientv3.WithPrefix(), clientv3.WithRev(start))
		for ev := range events {
			outputEvents := make([]*api.Volume, 0, len(ev.Events))

			for _, e := range ev.Events {
				decoded := &api.Volume{}
				err := proto.Unmarshal(e.Kv.Value, decoded)
				if err == nil {
					if e.Type == mvccpb.DELETE {
						decoded.State = api.VolumeState_DELETED
					}
					outputEvents = append(outputEvents, decoded)
				}
			}

			output <- outputEvents
		}

		close(output)
	}()

	return output
}

func (m *EtcdManager) ModifyVolume(ctx context.Context, name string, shards int64) error {
	// TODO: transactions or mutex
	key := "volumes/" + name
	resp, err := m.client.Get(ctx, key)
	if err != nil {
		return err
	}

	var vol *api.Volume

	if len(resp.Kvs) == 0 {
		if shards <= 0 {
			return nil
		}

		vol = &api.Volume{
			Name:            name,
			RequestedShards: shards,
			Shards:          0,
			State:           api.VolumeState_RESIZING,
		}
	} else {
		vol = &api.Volume{}
		err = proto.Unmarshal(resp.Kvs[0].Value, vol)
		if err != nil {
			return err
		}

		if shards == vol.RequestedShards {
			return nil
		}

		vol.RequestedShards = shards
		vol.State = api.VolumeState_RESIZING
	}

	bytes, err := proto.Marshal(vol)
	if err != nil {
		return err
	}
	_, err = m.client.Put(ctx, key, string(bytes))
	return err
}

func (m *EtcdManager) UpdateVolumeShards(ctx context.Context, name string, added []int64, removed []int64) error {
	// TODO: transactions or mutex
	key := "volumes/" + name
	resp, err := m.client.Get(ctx, key)
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return errors.New("volume not found")
	}

	vol := &api.Volume{}
	err = proto.Unmarshal(resp.Kvs[0].Value, vol)
	if err != nil {
		return err
	}

	currentShardsMap := make(map[int64]bool)
	for _, v := range vol.CurrentShards {
		currentShardsMap[v] = true
	}
	for _, v := range added {
		currentShardsMap[v] = true
	}
	for _, v := range removed {
		delete(currentShardsMap, v)
	}
	currentShards := make([]int64, 0, len(currentShardsMap))
	for shard := range currentShardsMap {
		currentShards = append(currentShards, shard)
	}

	vol.CurrentShards = currentShards
	vol.Shards = int64(len(currentShards))

	if vol.Shards == 0 && vol.RequestedShards == 0 {
		_, err = m.client.Delete(ctx, key)
	} else {
		if vol.Shards == vol.RequestedShards {
			vol.State = api.VolumeState_READY
		} else {
			vol.State = api.VolumeState_RESIZING
		}

		var bytes []byte
		bytes, err = proto.Marshal(vol)
		if err != nil {
			return err
		}
		_, err = m.client.Put(ctx, key, string(bytes))
	}

	return err
}

func (m *EtcdManager) fetchAndUnmarshal(ctx context.Context, key string, msg proto.Message) (int64, error) {
	resp, err := m.client.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if len(resp.Kvs) == 0 {
		return 0, errors.New("key not found")
	}

	err = proto.Unmarshal(resp.Kvs[0].Value, msg)
	if err != nil {
		return 0, err
	}

	return resp.Kvs[0].ModRevision, nil
}

func (m *EtcdManager) initMonitor(ctx context.Context, boot BootstrapData) error {
	cs := &api.ClientSettings{
		MinFillSize: boot.Settings.MinFillSize,
		HashAlgo:    api.HashAlgorithm_SHA256,
		ChunkAlgo: &api.ClientSettings_Rabin{
			Rabin: &api.RabinChunkAlgorithm{
				MaxSize: boot.Settings.RabinMaxSize,
				MinSize: boot.Settings.RabinMinSize,
				Poly:    boot.Settings.RabinPoly,
			},
		},
	}

	ss := &api.ServerSettings{
		ShardEntriesPerSlice: boot.Settings.ShardEntriesPerSlice,
		ShardSize:            boot.Settings.ShardSize,
	}

	blocks := &api.NodeReplicaSets{Sets: make([]*api.NodeReplicaSet, 0, 1)}
	indices := &api.NodeReplicaSets{Sets: make([]*api.NodeReplicaSet, 0, 1)}

	for _, block := range boot.Blocks {
		blocks.Sets = append(blocks.Sets, &api.NodeReplicaSet{
			Nodes: []*api.Node{
				&api.Node{
					Address: block.Address,
					Name:    block.Name,
					Primary: true,
				},
			},
		})
	}

	for _, idx := range boot.Indexes {
		indices.Sets = append(indices.Sets, &api.NodeReplicaSet{
			Nodes: []*api.Node{
				&api.Node{
					Address: idx.Address,
					Name:    idx.Name,
					Primary: true,
				},
			},
		})
	}

	volume := &api.Volume{
		Name:            "first",
		RequestedShards: 1,
		Shards:          0,
		State:           api.VolumeState_RESIZING,
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
