package main

import (
	"git.eplight.org/eplightning/ddfs/pkg/api"
	"os"

	"git.eplight.org/eplightning/ddfs/pkg/chunker"
	"git.eplight.org/eplightning/ddfs/pkg/index"
	"github.com/davecgh/go-spew/spew"
	"github.com/dgraph-io/badger"
	"github.com/hashicorp/golang-lru"
)

func main() {
	shardTest()
}

func shardTest() {
	
	opts := badger.DefaultOptions
	opts.Dir = "S:\\badger"
	opts.ValueDir = "S:\\badger"

	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	cache, err := lru.NewARC(1024)
	if err != nil {
		panic(err)
	}

	shard, err := index.NewShard(index.ShardSettings{
		ID: "id",
		Size: 512 * 1024 * 1024,
		SliceSize: 3,
		New: false,
	}, db, cache)
	if err != nil {
		panic(err)
	}

	// TODO: End should be inclusive, currently is exclusive
	data, err := shard.GetRange(0, 99999)
	spew.Dump(data, err)

	for _, x := range data.Entries {
		switch e := x.Entry.(type) {
		case (*api.IndexEntry_Fill):
			spew.Dump("fill", e.Fill.Byte)
		case (*api.IndexEntry_Hash):
			spew.Dump("hash", e.Hash.Hash)
		default:
			spew.Dump("wtf")
		}
	}

	newEntries := []*api.IndexEntry{
		createFillEntry(1000, 0),
		createFillEntry(1000, 1),
		createFillEntry(1000, 2),
		createFillEntry(1000, 3),
		createFillEntry(1000, 4),
		createFillEntry(1000, 5),
		createFillEntry(512 * 1024 * 1024 - 6000, 6),
	}

	err = shard.PutRange(index.RangeData{
		Start: 0,
		End: 512 * 1024 * 1024,
		Entries: newEntries,
	})

	spew.Dump(err)

}

func chunkTest() {
	file, _ := os.Open("S:\\test.txt")

	baseChunker := chunker.NewRabinChunker(512, 1024, 1241242412)
	fillChunker := chunker.NewFillDetectingChunker(baseChunker, 50)

	fillChunker.Reset(file)

	var err error
	var chunk chunker.Chunk

	for err == nil {
		chunk, err = fillChunker.NextChunk()

		spew.Dump(chunk)
	}
}

func createFillEntry(size int64, b int32) *api.IndexEntry {
	return &api.IndexEntry{
		BlockSize: size,
		Entry: &api.IndexEntry_Fill{
			Fill: &api.FillIndexEntry{
				Byte: b,
			},
		},
	}
}

func createHashEntry(size int64, hash []byte) *api.IndexEntry {
	return &api.IndexEntry{
		BlockSize: size,
		Entry: &api.IndexEntry_Hash{
			Hash: &api.HashIndexEntry{
				Hash: hash,
			},
		},
	}
}
