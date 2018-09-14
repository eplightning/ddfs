package client

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"git.eplight.org/eplightning/ddfs/pkg/api"
	"git.eplight.org/eplightning/ddfs/pkg/monitor"
	"git.eplight.org/eplightning/ddfs/pkg/util"
	"google.golang.org/grpc"
	tomb "gopkg.in/tomb.v2"
)

type HashedData struct {
	Hash []byte
	Data []byte
}

type grpcBlockClientConnection struct {
	api.BlockStoreClient
	io.Closer
}

type blockChunk struct {
	node   string
	chunks [][]byte
}

type reserveChunk struct {
	node   string
	chunks []*HashedData
}

type BlockClient struct {
	ring        util.HashRing
	connections map[string]*grpcBlockClientConnection
	cs          *api.ClientSettings
	connLock    sync.Mutex
}

func NewBlockClient(ctx context.Context, mon monitor.Client) (*BlockClient, error) {
	cs, err := mon.GetClientSettings(ctx, &api.ClientSettingsRequest{})
	if err != nil {
		return nil, err
	}
	nodes, err := mon.GetBlockStores(ctx, &api.GetBlockStoresRequest{})
	if err != nil {
		return nil, err
	}

	return &BlockClient{
		ring:        util.NewHashRing(nodes.Data),
		connections: make(map[string]*grpcBlockClientConnection),
		cs:          cs.Settings,
	}, nil
}

func (cl *BlockClient) WriteBlocks(ctx context.Context, blocks []*HashedData) error {
	chunks, err := cl.splitWorkReserve(blocks)
	if err != nil {
		return err
	}

	tracker, subCtx := tomb.WithContext(ctx)
	for _, chunk := range chunks {
		func() {
			chunk := chunk
			tracker.Go(func() error {
				err := cl.putAndReserve(subCtx, chunk.node, chunk.chunks)
				if err != nil {
					return err
				}
				return nil
			})
		}()
	}

	return tracker.Wait()
}

func (cl *BlockClient) GetBlock(ctx context.Context, hash []byte) ([]byte, error) {
	node := cl.ring.Block(util.NewBlockHash(hash))
	if node == nil {
		return nil, errors.New("could not find node for block")
	}
	r, err := cl.retrieveBlocks(ctx, node.Name, [][]byte{hash})
	if err != nil {
		return nil, err
	}
	stringHash := string(hash)
	x, ok := r[stringHash]
	if !ok {
		return nil, errors.New("not found")
	}
	return x, nil
}

func (cl *BlockClient) GetBlocks(ctx context.Context, indices []*RangeItem) (map[string][]byte, error) {
	chunks, err := cl.splitWork(indices)
	if err != nil {
		return nil, err
	}

	output := make(map[string][]byte)
	if len(chunks) == 0 {
		return output, nil
	}

	mergeCh := make(chan map[string][]byte, 10)
	mergeWait := make(chan struct{})

	go func() {
		for m := range mergeCh {
			for k, v := range m {
				output[k] = v
			}
		}
		close(mergeWait)
	}()

	tracker, subCtx := tomb.WithContext(ctx)
	for _, chunk := range chunks {
		func() {
			chunk := chunk
			tracker.Go(func() error {
				r, err := cl.retrieveBlocks(subCtx, chunk.node, chunk.chunks)
				if err != nil {
					return err
				}
				mergeCh <- r
				return nil
			})
		}()
	}

	err = tracker.Wait()
	if err != nil {
		return nil, err
	}

	close(mergeCh)
	<-mergeWait

	return output, nil

}

func (cl *BlockClient) splitWorkReserve(blocks []*HashedData) ([]*reserveChunk, error) {
	// 400 000
	byNode := make(map[string][]*HashedData)
	chunks := make([]*reserveChunk, 0, 5000)

	for _, entry := range blocks {
		node := cl.ring.Block(util.NewBlockHash(entry.Hash))
		if node == nil {
			return nil, errors.New("could not find node for block")
		}

		nodeBlocks, ok := byNode[node.Name]
		if !ok {
			nodeBlocks = make([]*HashedData, 0, 1000)
			byNode[node.Name] = nodeBlocks
		} else {
			if len(nodeBlocks) > 400000 {
				chunks = append(chunks, &reserveChunk{
					chunks: nodeBlocks,
					node:   node.Name,
				})
				nodeBlocks = make([]*HashedData, 0, 1000)
				byNode[node.Name] = nodeBlocks
			}
		}

		byNode[node.Name] = append(nodeBlocks, entry)
	}

	for k, v := range byNode {
		if len(v) > 0 {
			chunks = append(chunks, &reserveChunk{
				chunks: v,
				node:   k,
			})
		}
	}

	return chunks, nil
}

func (cl *BlockClient) splitWork(indices []*RangeItem) ([]*blockChunk, error) {
	// 400 000
	byNode := make(map[string]([][]byte))
	chunks := make([]*blockChunk, 0, 5000)

	for _, idx := range indices {
		for _, x := range idx.Slices {
			for _, entry := range x.Entries {
				h, ok := entry.Entry.(*api.IndexEntry_Hash)
				if ok {
					node := cl.ring.Block(util.NewBlockHash(h.Hash.Hash))
					if node == nil {
						return nil, errors.New("could not find node for block")
					}

					nodeBlocks, ok := byNode[node.Name]
					if !ok {
						nodeBlocks = make([][]byte, 0, 1000)
						byNode[node.Name] = nodeBlocks
					} else {
						if len(nodeBlocks) > 400000 {
							chunks = append(chunks, &blockChunk{
								chunks: nodeBlocks,
								node:   node.Name,
							})
							nodeBlocks = make([][]byte, 0, 1000)
							byNode[node.Name] = nodeBlocks
						}
					}

					byNode[node.Name] = append(nodeBlocks, h.Hash.Hash)
				}
			}
		}
	}

	for k, v := range byNode {
		if len(v) > 0 {
			chunks = append(chunks, &blockChunk{
				chunks: v,
				node:   k,
			})
		}
	}

	return chunks, nil
}

func (cl *BlockClient) connection(name string) (*grpcBlockClientConnection, error) {
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

	conn := &grpcBlockClientConnection{
		Closer:           cc,
		BlockStoreClient: api.NewBlockStoreClient(cc),
	}
	cl.connections[name] = conn
	return conn, nil
}

func (cl *BlockClient) putAndReserve(ctx context.Context, node string, chunks []*HashedData) error {
	conn, err := cl.connection(node)
	if err != nil {
		return err
	}

	hashes := make([][]byte, len(chunks))
	for i, data := range chunks {
		hashes[i] = data.Hash
	}

	resp, err := conn.Reserve(ctx, &api.BlockReserveRequest{
		Hashes: hashes,
	})
	if err != nil {
		return err
	}
	if resp.Header.Error != 0 {
		return errors.New(resp.Header.ErrorMsg)
	}

	stream, err := conn.Put(ctx)
	if err != nil {
		return err
	}

	err = stream.Send(&api.BlockPutRequest{
		Msg: &api.BlockPutRequest_Info_{
			Info: &api.BlockPutRequest_Info{
				ReservationId: resp.ReservationId,
			},
		},
	})
	if err != nil {
		return err
	}

	chunk := make([][]byte, 0, 5)
	maxLen := util.MaxMessageSize - 4096
	var lenny int

	for _, missing := range resp.MissingBlocks {
		c := chunks[missing]
		nextLen := lenny + len(c.Data)
		if nextLen > maxLen && len(chunk) > 0 {
			stream.Send(&api.BlockPutRequest{
				Msg: &api.BlockPutRequest_Data_{
					Data: &api.BlockPutRequest_Data{
						Blocks: chunk,
					},
				},
			})
			chunk = make([][]byte, 0, 5)
		}
		lenny = nextLen
		chunk = append(chunk, c.Data)
	}
	if len(chunk) > 0 {
		stream.Send(&api.BlockPutRequest{
			Msg: &api.BlockPutRequest_Data_{
				Data: &api.BlockPutRequest_Data{
					Blocks: chunk,
				},
			},
		})
	}

	r, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	if r.Header.Error != 0 {
		return errors.New(r.Header.ErrorMsg)
	}

	return nil
}

func (cl *BlockClient) retrieveBlocks(ctx context.Context, node string, hashes [][]byte) (map[string][]byte, error) {
	conn, err := cl.connection(node)
	if err != nil {
		return nil, err
	}
	stream, err := conn.Get(ctx, &api.BlockGetRequest{
		Hashes: hashes,
	})
	if err != nil {
		return nil, err
	}

	output := make(map[string][]byte)
	var ptr int
	for {
		dataMsg, err := stream.Recv()
		if err != nil {
			break
		}
		data, ok := dataMsg.Msg.(*api.BlockGetResponse_Data_)
		if !ok {
			info, ok := dataMsg.Msg.(*api.BlockGetResponse_Info_)
			if ok {
				if info.Info.Header.Error != 0 {
					return nil, errors.New(info.Info.Header.ErrorMsg)
				}
			}
			break
		}
		for _, b := range data.Data.Blocks {
			output[string(hashes[ptr])] = b
			ptr++
		}
	}

	return output, nil
}
