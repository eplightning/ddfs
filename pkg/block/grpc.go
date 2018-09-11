package block

import (
	"context"

	"git.eplight.org/eplightning/ddfs/pkg/api"
	"git.eplight.org/eplightning/ddfs/pkg/util"
	"github.com/pkg/errors"
)

type BlockGrpc struct {
	block *BlockManager
}

func NewBlockGrpc(block *BlockManager) *BlockGrpc {
	return &BlockGrpc{
		block: block,
	}
}

func (s *BlockGrpc) Get(r *api.BlockGetRequest, stream api.BlockStore_GetServer) error {
	ch := s.block.GetChunked(r.Hashes, util.MaxMessageSize-4096)

	for chunk := range ch {
		if chunk == nil {
			stream.Send(&api.BlockGetResponse{
				Msg: &api.BlockGetResponse_Info_{
					Info: &api.BlockGetResponse_Info{
						Header: s.buildHeader(errors.New("data not found")),
					},
				},
			})
			return nil
		}
		err := stream.Send(&api.BlockGetResponse{
			Msg: &api.BlockGetResponse_Data_{
				Data: &api.BlockGetResponse_Data{
					Blocks: chunk,
				},
			},
		})
		if err != nil {
			return err
		}
	}

	err := stream.Send(&api.BlockGetResponse{
		Msg: &api.BlockGetResponse_Info_{
			Info: &api.BlockGetResponse_Info{
				Header: s.buildHeader(nil),
			},
		},
	})

	return err
}

func (s *BlockGrpc) Put(stream api.BlockStore_PutServer) error {
	first, err := stream.Recv()
	if err != nil {
		return stream.SendAndClose(&api.BlockPutResponse{
			Header: s.buildHeader(err),
		})
	}

	firstMsg, ok := first.Msg.(*api.BlockPutRequest_Info_)
	if !ok {
		return stream.SendAndClose(&api.BlockPutResponse{
			Header: s.buildHeader(errors.New("invalid request sequence")),
		})
	}

	ch := make(chan [][]byte, 100)
	end := make(chan struct{})
	defer close(end)

	go func() {
		defer close(ch)

		for {
			x, err := stream.Recv()
			if err != nil {
				return
			}
			dataMsg, ok := x.Msg.(*api.BlockPutRequest_Data_)
			if !ok {
				return
			}

			select {
			case ch <- dataMsg.Data.Blocks:
			case <-end:
				return
			}
		}
	}()

	err = s.block.Put(firstMsg.Info.ReservationId, ch)
	return stream.SendAndClose(&api.BlockPutResponse{
		Header: s.buildHeader(err),
	})
}

func (s *BlockGrpc) Reserve(ctx context.Context, r *api.BlockReserveRequest) (*api.BlockReserveResponse, error) {
	id, missing, err := s.block.Reserve(r.Hashes)
	return &api.BlockReserveResponse{
		Header:        s.buildHeader(err),
		MissingBlocks: missing,
		ReservationId: id,
	}, nil
}

func (s *BlockGrpc) Delete(ctx context.Context, r *api.BlockDeleteRequest) (*api.BlockDeleteResponse, error) {
	err := s.block.Delete(r.Hashes)
	return &api.BlockDeleteResponse{
		Header: s.buildHeader(err),
	}, nil
}

func (s *BlockGrpc) buildHeader(err error) *api.BlockResponseHeader {
	var errCode int32
	var errMsg string
	if err != nil {
		errCode = 1
		errMsg = err.Error()
	}

	return &api.BlockResponseHeader{
		Error:    errCode,
		ErrorMsg: errMsg,
	}
}
