package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"git.eplight.org/eplightning/ddfs/pkg/api"
	"git.eplight.org/eplightning/ddfs/pkg/monitor"
	"git.eplight.org/eplightning/ddfs/pkg/util"
	"google.golang.org/grpc"
)

type grpcIndexClientConnection struct {
	api.IndexStoreClient
	io.Closer
}

type ShardBoundary struct {
	Start int64
	End   int64
}

type RangeResponse struct {
	Start  int64
	End    int64
	Slices []*api.IndexSlice
	Bounds []ShardBoundary
}

type IndexClient struct {
	ring        util.HashRing
	connections map[string]*grpcIndexClientConnection
	cs          *api.ClientSettings
	ss          *api.ServerSettings
}

func NewIndexClient(ctx context.Context, mon monitor.Client) (*IndexClient, error) {
	cs, err := mon.GetClientSettings(ctx, &api.ClientSettingsRequest{})
	if err != nil {
		return nil, err
	}
	ss, err := mon.GetServerSettings(ctx, &api.ServerSettingsRequest{})
	if err != nil {
		return nil, err
	}
	nodes, err := mon.GetIndexStores(ctx, &api.GetIndexStoresRequest{})
	if err != nil {
		return nil, err
	}

	return &IndexClient{
		ring:        util.NewHashRing(nodes.Data),
		connections: make(map[string]*grpcIndexClientConnection),
		cs:          cs.Settings,
		ss:          ss.Settings,
	}, nil
}

func (cl *IndexClient) GetRange(ctx context.Context, volume string, start, end int64) (*RangeResponse, error) {
	log.Println("Volume ", volume, start, end)
	s := cl.shard(volume, start)
	log.Println("shard ", s)
	node := cl.ring.Shard(s)
	log.Println("node ", node)
	conn, err := cl.connection(node.Name)
	if err != nil {
		return nil, err
	}
	log.Printf("conn %v %v\n", conn, err)
	stream, err := conn.GetRange(ctx, &api.IndexGetRangeRequest{
		End:   end,
		Shard: s,
		Start: start,
	})
	if err != nil {
		return nil, err
	}

	firstMsg, err := stream.Recv()
	if err != nil {
		return nil, err
	}
	first, ok := firstMsg.Msg.(*api.IndexGetRangeResponse_Info_)
	if !ok {
		return nil, errors.New("invalid first message")
	}

	log.Printf("first %v", first.Info)

	for {
		dataMsg, err := stream.Recv()
		if err != nil {
			break
		}
		data, ok := dataMsg.Msg.(*api.IndexGetRangeResponse_Data_)
		if !ok {
			break
		}
		fmt.Printf("data %v", data)
	}

	err = stream.CloseSend()
	if err != nil {
		panic(err)
	}

	return nil, nil
}

func (cl *IndexClient) connection(name string) (*grpcIndexClientConnection, error) {
	c, ok := cl.connections[name]
	if ok {
		return c, nil
	}

	node := cl.ring.Node(name)
	if node == nil {
		return nil, errors.New("unknown node")
	}

	cc, err := grpc.Dial(
		node.Address, grpc.WithTimeout(15*time.Second), grpc.WithInsecure(),
		grpc.WithMaxMsgSize(util.MaxMessageSize),
	)
	if err != nil {
		return nil, errors.New("cannot connect via gRPC")
	}

	conn := &grpcIndexClientConnection{
		Closer:           cc,
		IndexStoreClient: api.NewIndexStoreClient(cc),
	}
	cl.connections[name] = conn
	return conn, nil
}

func (cl *IndexClient) shard(volume string, start int64) string {
	idx := start / cl.ss.ShardSize
	return volume + "-" + strconv.FormatInt(idx, 10)
}
