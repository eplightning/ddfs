package monitor

import (
	"context"

	"git.eplight.org/eplightning/ddfs/pkg/api"
)

type MonitorGrpc struct {
	etcd *EtcdManager
}

func NewMonitorGrpc(etcd *EtcdManager) *MonitorGrpc {
	return &MonitorGrpc{
		etcd: etcd,
	}
}

func (s *MonitorGrpc) GetClientSettings(ctx context.Context, r *api.ClientSettingsRequest) (*api.ClientSettingsResponse, error) {
	result, rev, err := s.etcd.ClientSettings(ctx)

	response := &api.ClientSettingsResponse{
		Header:   s.buildHeader(err, rev),
		Settings: result,
	}
	return response, nil
}

func (s *MonitorGrpc) GetServerSettings(ctx context.Context, r *api.ServerSettingsRequest) (*api.ServerSettingsResponse, error) {
	result, rev, err := s.etcd.ServerSettings(ctx)

	response := &api.ServerSettingsResponse{
		Header:   s.buildHeader(err, rev),
		Settings: result,
	}
	return response, nil
}

func (s *MonitorGrpc) GetBlockStores(ctx context.Context, r *api.GetBlockStoresRequest) (*api.GetBlockStoresResponse, error) {
	result, rev, err := s.etcd.BlockStores(ctx)

	response := &api.GetBlockStoresResponse{
		Header: s.buildHeader(err, rev),
		Data:   result,
	}
	return response, nil
}

func (s *MonitorGrpc) GetIndexStores(ctx context.Context, r *api.GetIndexStoresRequest) (*api.GetIndexStoresResponse, error) {
	result, rev, err := s.etcd.IndexStores(ctx)

	response := &api.GetIndexStoresResponse{
		Header: s.buildHeader(err, rev),
		Data:   result,
	}
	return response, nil
}

func (s *MonitorGrpc) Get(ctx context.Context, r *api.GetVolumesRequest) (*api.GetVolumesResponse, error) {
	result, rev, err := s.etcd.Volumes(ctx)

	response := &api.GetVolumesResponse{
		Header:  s.buildHeader(err, rev),
		Volumes: result,
	}

	return response, nil
}

func (s *MonitorGrpc) Modify(ctx context.Context, r *api.ModifyVolumeRequest) (*api.ModifyVolumeResponse, error) {
	err := s.etcd.ModifyVolume(ctx, r.Name, r.Shards)

	response := &api.ModifyVolumeResponse{
		Header: s.buildHeader(err, 0),
	}

	return response, nil
}

func (s *MonitorGrpc) UpdateShards(ctx context.Context, r *api.UpdateShardsRequest) (*api.UpdateShardsResponse, error) {
	err := s.etcd.UpdateVolumeShards(ctx, r.VolumeName, r.AddedShards, r.RemovedShards)

	response := &api.UpdateShardsResponse{
		Header: s.buildHeader(err, 0),
	}

	return response, nil
}

func (s *MonitorGrpc) Watch(stream api.Volumes_WatchServer) error {
	msg, err := stream.Recv()
	if err != nil {
		return err
	}
	// ok?
	if msg.Header.End {
		return nil
	}

	watchContext, cancel := context.WithCancel(stream.Context())
	defer cancel()
	go func() {
		defer cancel()
		stream.Recv()
	}()

	watch := s.etcd.WatchVolumes(watchContext, msg.Header.StartRevision)

	for w := range watch {
		err = stream.Send(&api.WatchVolumesResponse{
			Header:  s.buildHeader(nil, 0),
			Volumes: w,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MonitorGrpc) buildHeader(err error, rev int64) *api.MonitorResponseHeader {
	var errCode int32
	var errMsg string
	if err != nil {
		errCode = 1
		errMsg = err.Error()
	}

	return &api.MonitorResponseHeader{
		Error:    errCode,
		ErrorMsg: errMsg,
		Revision: rev,
	}
}
