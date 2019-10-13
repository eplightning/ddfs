package monitor

import (
	"io"

	"github.com/eplightning/ddfs/pkg/api"
	"google.golang.org/grpc"
)

type Client interface {
	api.NodesClient
	api.VolumesClient
	api.SettingsClient
	io.Closer
}

type CombinedClient struct {
	api.NodesClient
	api.VolumesClient
	api.SettingsClient
	io.Closer
}

func FromClientConn(cc *grpc.ClientConn) Client {
	return &CombinedClient{
		NodesClient:    api.NewNodesClient(cc),
		VolumesClient:  api.NewVolumesClient(cc),
		SettingsClient: api.NewSettingsClient(cc),
		Closer:         cc,
	}
}
