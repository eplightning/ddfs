package index

import "git.eplight.org/eplightning/ddfs/pkg/api"

type IndexGrpc struct {
	shards *ShardManager
}

func NewIndexGrpc(shards *ShardManager) *IndexGrpc {
	return &IndexGrpc{
		shards: shards,
	}
}

func (s *IndexGrpc) GetRange(r *api.IndexGetRangeRequest, stream api.IndexStore_GetRangeServer) error {
	shard, err := s.shards.Shard(r.Shard)
	if err != nil {
		err = stream.Send(&api.IndexGetRangeResponse{
			Msg: &api.IndexGetRangeResponse_Info_{
				Info: &api.IndexGetRangeResponse_Info{
					Header: s.buildHeader(err, 0),
					Start:  0,
					End:    0,
				},
			},
		})
		return err
	}

	data, err := shard.GetRange(r.Start, r.End)
	if err != nil {
		err = stream.Send(&api.IndexGetRangeResponse{
			Msg: &api.IndexGetRangeResponse_Info_{
				Info: &api.IndexGetRangeResponse_Info{
					Header: s.buildHeader(err, 0),
					Start:  0,
					End:    0,
				},
			},
		})
		return err
	}

	err = stream.Send(&api.IndexGetRangeResponse{
		Msg: &api.IndexGetRangeResponse_Info_{
			Info: &api.IndexGetRangeResponse_Info{
				Header: s.buildHeader(nil, 0),
				Start:  data.Start,
				End:    data.End,
			},
		},
	})
	if err != nil {
		return err
	}

	// TODO: slice it up
	err = stream.Send(&api.IndexGetRangeResponse{
		Msg: &api.IndexGetRangeResponse_Data_{
			Data: &api.IndexGetRangeResponse_Data{
				Slice: &api.IndexSlice{
					Entries: data.Entries,
				},
			},
		},
	})

	return err
}

func (s *IndexGrpc) PutRange(stream api.IndexStore_PutRangeServer) error {
	// TODO: todo
	return nil
}

func (s *IndexGrpc) buildHeader(err error, rev int64) *api.IndexResponseHeader {
	var errCode int32
	var errMsg string
	if err != nil {
		errCode = 1
		errMsg = err.Error()
	}

	return &api.IndexResponseHeader{
		Error:    errCode,
		ErrorMsg: errMsg,
		Revision: rev,
	}
}
