package block

import (
	"context"

	"git.eplight.org/eplightning/ddfs/pkg/api"
)

type BlockGrpc struct {
}

func NewBlockGrpc() *BlockGrpc {
	return &BlockGrpc{}
}

func (s *BlockGrpc) Get(r *api.BlockGetRequest, stream api.BlockStore_GetServer) error {
	return nil
}

func (s *BlockGrpc) Put(stream api.BlockStore_PutServer) error {
	return nil
}

func (s *BlockGrpc) Reserve(ctx context.Context, r *api.BlockReserveRequest) (*api.BlockReserveResponse, error) {
	return nil, nil
}

func (s *BlockGrpc) Delete(ctx context.Context, r *api.BlockDeleteRequest) (*api.BlockDeleteResponse, error) {
	return nil, nil
}
