package client

import (
	"context"
	"errors"
	"io"
	"strconv"
	"sync"
	"time"

	"git.eplight.org/eplightning/ddfs/pkg/api"
	"git.eplight.org/eplightning/ddfs/pkg/monitor"
	"git.eplight.org/eplightning/ddfs/pkg/util"
	"google.golang.org/grpc"
	tomb "gopkg.in/tomb.v2"
)

type grpcIndexClientConnection struct {
	api.IndexStoreClient
	io.Closer
}

type shardBoundary struct {
	Start int64
	End   int64
	Index int
}

type RangeResponse struct {
	Ranges []*RangeItem
}

type RangeItem struct {
	Start  int64
	End    int64
	Slices []*api.IndexSlice
}

type IndexClient struct {
	ring        util.HashRing
	connections map[string]*grpcIndexClientConnection
	cs          *api.ClientSettings
	ss          *api.ServerSettings
	connLock    sync.Mutex
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

func (cl *IndexClient) GetRange(ctx context.Context, volume string, start, end int64) ([]*RangeItem, error) {
	// transform to offsets local to shards
	bounds := cl.getBounds(start, end)

	ranges := make([]*RangeItem, len(bounds))
	tracker, subCtx := tomb.WithContext(ctx)

	for i, bound := range bounds {
		func() {
			bound := bound
			i := i
			tracker.Go(func() error {
				r, err := cl.retrieveShard(subCtx, volume, bound)
				if err != nil {
					return err
				}
				ranges[i] = r
				return nil
			})
		}()
	}

	err := tracker.Wait()
	if err != nil {
		return nil, err
	}

	// transform to global offsets
	for i, r := range ranges {
		diff := int64(bounds[i].Index) * cl.ss.ShardSize
		r.Start += diff
		r.End += diff
	}

	return ranges, nil
}

func (cl *IndexClient) PutShard(ctx context.Context, volume string, start, end int64, entries []*api.IndexEntry) error {
	bounds := cl.getBounds(start, end)
	if len(bounds) != 1 {
		return errors.New("index data out of shard bounds")
	}
	boundary := bounds[0]

	s := cl.shardName(volume, boundary.Index)
	node := cl.ring.Shard(s)
	conn, err := cl.connection(node.Name)
	if err != nil {
		return err
	}
	stream, err := conn.PutRange(ctx)
	if err != nil {
		return err
	}

	err = stream.Send(&api.IndexPutRangeRequest{
		Msg: &api.IndexPutRangeRequest_Info_{
			Info: &api.IndexPutRangeRequest_Info{
				Start: start,
				End:   end,
				Shard: s,
			},
		},
	})
	if err != nil {
		return err
	}

	for i := 0; i < len(entries); {
		pack := len(entries) - i
		if pack > 360000 {
			pack = 360000
		}

		err = stream.Send(&api.IndexPutRangeRequest{
			Msg: &api.IndexPutRangeRequest_Data_{
				Data: &api.IndexPutRangeRequest_Data{
					Slice: &api.IndexSlice{
						Entries: entries[i : i+pack],
					},
				},
			},
		})
		if err != nil {
			return err
		}

		i += pack
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	if resp.Header.Error != 0 {
		return errors.New(resp.Header.ErrorMsg)
	}

	return nil
}

func (cl *IndexClient) connection(name string) (*grpcIndexClientConnection, error) {
	cl.connLock.Lock()
	defer cl.connLock.Unlock()
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

func (cl *IndexClient) retrieveShard(ctx context.Context, volume string, boundary shardBoundary) (*RangeItem, error) {
	s := cl.shardName(volume, boundary.Index)
	node := cl.ring.Shard(s)
	conn, err := cl.connection(node.Name)
	if err != nil {
		return nil, err
	}
	stream, err := conn.GetRange(ctx, &api.IndexGetRangeRequest{
		End:   boundary.End,
		Shard: s,
		Start: boundary.Start,
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

	output := &RangeItem{
		End:    first.Info.End,
		Start:  first.Info.Start,
		Slices: nil,
	}

	for {
		dataMsg, err := stream.Recv()
		if err != nil {
			break
		}
		data, ok := dataMsg.Msg.(*api.IndexGetRangeResponse_Data_)
		if !ok {
			break
		}
		output.Slices = append(output.Slices, data.Data.Slice)
	}

	return output, nil
}

func (cl *IndexClient) shardName(volume string, idx int) string {
	return volume + "-" + strconv.Itoa(idx)
}

func (cl *IndexClient) getBounds(start, end int64) []shardBoundary {
	first := int(start / cl.ss.ShardSize)
	second := int((end - 1) / cl.ss.ShardSize)
	count := second - first + 1

	bounds := make([]shardBoundary, count)
	bounds[0].Start = start - int64(first)*cl.ss.ShardSize
	bounds[count-1].End = end - int64(second)*cl.ss.ShardSize

	for i := 0; i < count; i++ {
		if i != count-1 {
			bounds[i].End = cl.ss.ShardSize
		}
		bounds[i].Index = first + i
	}
	return bounds
}
