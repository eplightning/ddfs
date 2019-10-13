package client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"hash"
	"io"

	"golang.org/x/crypto/blake2b"

	"github.com/eplightning/ddfs/pkg/api"
	"github.com/eplightning/ddfs/pkg/chunker"
)

type CombinedClient struct {
	blk *BlockClient
	idx *IndexClient
	cs  *api.ClientSettings
}

func NewCombinedClient(block *BlockClient, index *IndexClient) *CombinedClient {
	return &CombinedClient{
		blk: block,
		idx: index,
		cs:  index.cs,
	}
}

func (c *CombinedClient) Read(ctx context.Context, volume string, start, end int64, dest io.Writer) error {
	ranges, err := c.idx.GetRange(ctx, volume, start, end)
	if err != nil {
		return err
	}

	blocks, err := c.blk.GetBlocks(ctx, ranges)
	if err != nil {
		return err
	}

	skip := start - ranges[0].Start

	readers := make([]io.Reader, 0, 100)
	for _, idx := range ranges {
		for _, x := range idx.Slices {
			for _, entry := range x.Entries {
				switch e := entry.Entry.(type) {
				case *api.IndexEntry_Hash:
					readers = append(readers, bytes.NewReader(
						blocks[string(e.Hash.Hash)],
					))
				case *api.IndexEntry_Fill:
					readers = append(readers, &IndexFillReader{
						Fill: byte(e.Fill.Byte),
						Size: entry.BlockSize,
					})
				}
			}
		}
	}

	if len(readers) == 0 {
		return nil
	}

	if skip > 0 {
		seeker := readers[0].(io.Seeker)
		seeker.Seek(skip, io.SeekCurrent)
	}

	multi := io.MultiReader(readers...)
	limited := io.LimitReader(multi, end-start)

	_, err = io.Copy(dest, limited)
	return err
}

func (c *CombinedClient) Write(ctx context.Context, volume string, start, end int64, source io.Reader) error {
	origStart := start
	origEnd := end

	ranges, err := c.idx.GetRange(ctx, volume, start, end)
	if err != nil {
		return err
	}

	start = ranges[0].Start
	end = ranges[len(ranges)-1].End

	last := len(ranges) - 1
	left := origStart
	var prepend, appendE *api.IndexEntry

	for i, shard := range ranges {
		prepend, appendE = nil, nil
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
					readers = append(readers, io.LimitReader(bytes.NewReader(data), origStart-start))
				} else {
					readers = append(readers, io.LimitReader(&IndexFillReader{Fill: byte(fill.Fill.Byte), Size: firstEntry.BlockSize}, origStart-start))
				}
			}
		}

		readers = append(readers, io.LimitReader(source, shard.End-left))
		left = shard.End

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
					r := bytes.NewReader(data)
					r.Seek(origEnd-end, io.SeekEnd)
					readers = append(readers, r)
				} else {
					r := &IndexFillReader{Fill: byte(fill.Fill.Byte), Size: lastEntry.BlockSize}
					r.Seek(origEnd-end, io.SeekEnd)
					readers = append(readers, r)
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

	var err error
	var baseChunker chunker.Chunker
	var hasher hash.Hash

	switch c.cs.HashAlgo {
	case api.HashAlgorithm_BLAKE2B256:
		hasher, err = blake2b.New256(nil)
		if err != nil {
			return nil, nil, err
		}
	case api.HashAlgorithm_SHA256:
		hasher = sha256.New()
	default:
		return nil, nil, errors.New("invalid chunker")
	}

	switch t := c.cs.ChunkAlgo.(type) {
	case *api.ClientSettings_Rabin:
		baseChunker = chunker.NewRabinChunker(int(t.Rabin.MinSize), int(t.Rabin.MaxSize), t.Rabin.Poly)
	case *api.ClientSettings_Fixed:
		baseChunker = chunker.NewFixedChunker(int(t.Fixed.MaxSize))
	default:
		return nil, nil, errors.New("invalid chunker")
	}

	fillChunker := chunker.NewFillDetectingChunker(baseChunker, int(c.cs.MinFillSize))
	fillChunker.Reset(reader)

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
