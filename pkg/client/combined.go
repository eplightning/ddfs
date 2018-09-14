package client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"

	"git.eplight.org/eplightning/ddfs/pkg/api"
	"git.eplight.org/eplightning/ddfs/pkg/chunker"
)

type CombinedClient struct {
	blk *BlockClient
	idx *IndexClient
}

func NewCombinedClient(block *BlockClient, index *IndexClient) *CombinedClient {
	return &CombinedClient{
		blk: block,
		idx: index,
	}
}

func (c *CombinedClient) Write(ctx context.Context, volume string, start, end int64, source io.Reader) error {
	origStart := start
	origEnd := end

	ranges, err := c.idx.GetRange(ctx, volume, start, end)
	if err != nil {
		return err
	}

	last := len(ranges) - 1
	for i, shard := range ranges {
		var prepend, appendE *api.IndexEntry
		readers := make([]io.Reader, 0, 1)

		if i == 0 && origStart > start {
			firstEntry := shard.Slices[0].Entries[0]
			fill, isFill := firstEntry.Entry.(*api.IndexEntry_Fill)

			if isFill && origStart-start >= int64(c.idx.cs.MinFillSize) {
				prepend = &api.IndexEntry{
					BlockSize: origStart - start,
					Entry:     fill,
				}
			} else {
				if !isFill {
					hash := firstEntry.Entry.(*api.IndexEntry_Hash)
					data, err := c.blk.GetBlock(ctx, hash.Hash.Hash)
					if err != nil {
						return err
					}
					readers = append(readers, bytes.NewReader(data))
				} else {
					readers = append(readers, &IndexFillReader{Fill: byte(fill.Fill.Byte), Size: firstEntry.BlockSize})
				}
			}
		}

		readers = append(readers, io.LimitReader(source, shard.End-shard.Start))

		if i == last && end > origEnd {
			lastSlice := shard.Slices[len(shard.Slices)-1]
			lastEntry := lastSlice.Entries[len(lastSlice.Entries)-1]
			fill, isFill := lastEntry.Entry.(*api.IndexEntry_Fill)

			if isFill && end-origEnd >= int64(c.idx.cs.MinFillSize) {
				appendE = &api.IndexEntry{
					BlockSize: end - origEnd,
					Entry:     fill,
				}
			} else {
				if !isFill {
					hash := lastEntry.Entry.(*api.IndexEntry_Hash)
					data, err := c.blk.GetBlock(ctx, hash.Hash.Hash)
					if err != nil {
						return err
					}
					readers = append(readers, bytes.NewReader(data))
				} else {
					readers = append(readers, &IndexFillReader{Fill: byte(fill.Fill.Byte), Size: lastEntry.BlockSize})
				}
			}
		}

		finalReader := io.MultiReader(readers...)

		entries, hashed, err := c.chunkIt(finalReader, prepend)
		if err != nil {
			return err
		}
		if appendE != nil {
			entries = append(entries, appendE)
		}

		err = c.blk.WriteBlocks(ctx, hashed)
		if err != nil {
			return err
		}

		err = c.idx.PutShard(ctx, volume, start, end, entries)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CombinedClient) chunkIt(reader io.Reader, pre *api.IndexEntry) ([]*api.IndexEntry, []*HashedData, error) {
	entries := make([]*api.IndexEntry, 0, 1000)

	if pre != nil {
		entries = append(entries, pre)
	}

	hashed := make([]*HashedData, 0, 1000)

	/*
		MaxSize: 4 * 1024 * 1024,
		MinSize: 100 * 1024,
		Poly:    12313278162312893,*/

	baseChunker := chunker.NewRabinChunker(100*1024, 4*1024*1024, 12313278162312893)
	fillChunker := chunker.NewFillDetectingChunker(baseChunker, int(c.idx.cs.MinFillSize))
	hasher := sha256.New()
	fillChunker.Reset(reader)

	var err error
	var chunk chunker.Chunk

	for {
		chunk, err = fillChunker.NextChunk()
		if err != nil {
			break
		}

		switch f := chunk.(type) {
		case *chunker.FillChunk:
			entries = append(entries, c.createFillEntry(int64(f.Size), int32(f.Fill)))
		case *chunker.RawChunk:
			hasher.Write(f.Raw)
			h := hasher.Sum(nil)
			hasher.Reset()
			hashed = append(hashed, &HashedData{
				Data: f.Raw,
				Hash: h,
			})

			entries = append(entries, c.createHashEntry(int64(len(f.Raw)), h))
		}
	}

	return entries, hashed, nil
}

func (c *CombinedClient) createFillEntry(size int64, b int32) *api.IndexEntry {
	return &api.IndexEntry{
		BlockSize: size,
		Entry: &api.IndexEntry_Fill{
			Fill: &api.FillIndexEntry{
				Byte: b,
			},
		},
	}
}

func (c *CombinedClient) createHashEntry(size int64, hash []byte) *api.IndexEntry {
	return &api.IndexEntry{
		BlockSize: size,
		Entry: &api.IndexEntry_Hash{
			Hash: &api.HashIndexEntry{
				Hash: hash,
			},
		},
	}
}
