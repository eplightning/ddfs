package index

import (
	"io"

	"git.eplight.org/eplightning/ddfs/pkg/api"
	"github.com/pkg/errors"
)

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

	for i := 0; i < len(data.Entries); {
		pack := len(data.Entries) - i
		if pack > 360000 {
			pack = 360000
		}

		err = stream.Send(&api.IndexGetRangeResponse{
			Msg: &api.IndexGetRangeResponse_Data_{
				Data: &api.IndexGetRangeResponse_Data{
					Slice: &api.IndexSlice{
						Entries: data.Entries[i : i+pack],
					},
				},
			},
		})
		if err != nil {
			return err
		}

		i += pack
	}

	return err
}

func (s *IndexGrpc) PutRange(stream api.IndexStore_PutRangeServer) error {
	first, err := stream.Recv()
	if err != nil {
		return s.sendPutResponse(err, stream)
	}

	firstMsg, ok := first.Msg.(*api.IndexPutRangeRequest_Info_)
	if !ok {
		return s.sendPutResponse(errors.New("invalid request sequence"), stream)
	}

	shard, err := s.shards.Shard(firstMsg.Info.Shard)
	if err != nil {
		return s.sendPutResponse(err, stream)
	}

	entries := make([]*api.IndexEntry, 0, 1)

	for {
		x, err := stream.Recv()
		if err != nil {
			break
		}

		dataMsg, ok := x.Msg.(*api.IndexPutRangeRequest_Data_)
		if !ok {
			err = errors.New("invalid request sequence")
			break
		}

		entries = append(entries, dataMsg.Data.Slice.Entries...)
	}
	if err != io.EOF {
		return s.sendPutResponse(err, stream)
	}

	err = shard.PutRange(RangeData{
		Start:   firstMsg.Info.Start,
		End:     firstMsg.Info.End,
		Entries: entries,
	})
	if err != nil {
		return s.sendPutResponse(err, stream)
	}

	return s.sendPutResponse(err, stream)
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

func (s *IndexGrpc) sendPutResponse(err error, stream api.IndexStore_PutRangeServer) error {
	return stream.SendAndClose(&api.IndexPutRangeResponse{
		Header: s.buildHeader(err, 0),
	})
}
