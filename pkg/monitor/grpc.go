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
	return nil, nil
}

func (s *MonitorGrpc) Get(ctx context.Context, r *api.GetVolumesRequest) (*api.GetVolumesResponse, error) {
	return nil, nil
}

func (s *MonitorGrpc) Create(ctx context.Context, r *api.CreateVolumeRequest) (*api.CreateVolumeResponse, error) {
	return nil, nil
}

func (s *MonitorGrpc) Delete(ctx context.Context, r *api.DeleteVolumeRequest) (*api.DeleteVolumeResponse, error) {
	return nil, nil
}

func (s *MonitorGrpc) Watch(stream api.Volumes_WatchServer) error {
	return nil
}

func (s *MonitorGrpc) UpdateShards(ctx context.Context, r *api.UpdateShardsRequest) (*api.UpdateShardsResponse, error) {
	return nil, nil
}

func (s *MonitorGrpc) GetBlockStores(ctx context.Context, r *api.GetBlockStoresRequest) (*api.GetBlockStoresResponse, error) {
	return nil, nil
}

func (s *MonitorGrpc) GetIndexStores(ctx context.Context, r *api.GetIndexStoresRequest) (*api.GetIndexStoresResponse, error) {
	return nil, nil
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
