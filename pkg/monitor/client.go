package monitor

import (
	"git.eplight.org/eplightning/ddfs/pkg/api"
	"google.golang.org/grpc"
)

type Client interface {
	api.NodesClient
	api.VolumesClient
	api.SettingsClient
}

type CombinedClient struct {
	api.NodesClient
	api.VolumesClient
	api.SettingsClient
}

func FromClientConn(cc *grpc.ClientConn) Client {
	return &CombinedClient{
		NodesClient:    api.NewNodesClient(cc),
		VolumesClient:  api.NewVolumesClient(cc),
		SettingsClient: api.NewSettingsClient(cc),
	}
}
