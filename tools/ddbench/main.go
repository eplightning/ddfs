package main

import (
	"crypto/sha256"
	"flag"
	"log"

	ch "github.com/eplightning/ddfs/pkg/chunker"
	"github.com/eplightning/ddfs/tools/ddbench/internal"
	"github.com/davecgh/go-spew/spew"
	"github.com/dgraph-io/badger"
)

var chunkerFlag = flag.String("chunker", "fixed", "chunking method [fixed|rabin]")
var chunkerFixedBlockSizeFlag = flag.Int("chunker-block-size", 16*4096, "block size for fixed chunker")
var chunkerVarMinSizeFlag = flag.Int("chunker-min-size", 512*1024, "min block size for variable chunkers")
var chunkerVarMaxSizeFlag = flag.Int("chunker-max-size", 8*1024*1024, "max block size for variable chunkers")
var dataDirectory = flag.String("data-path", "/tmp/data", "path to data")
var dbDirectory = flag.String("db-path", "/tmp/db", "where to store BadgerDB data")

func main() {
	flag.Parse()

	log.Println("Deduplication Benchmark")

	pathChannel := make(chan string, 1)
	hasher := sha256.New()

	var baseChunker ch.Chunker

	fillSize := 4096

	switch *chunkerFlag {
	case "rabin":
		baseChunker = ch.NewRabinChunker(*chunkerVarMinSizeFlag, *chunkerVarMaxSizeFlag, 14152035864944967)
		fillSize = *chunkerFixedBlockSizeFlag
		log.Println("Using rabin chunker")
	default:
		baseChunker = ch.NewFixedChunker(*chunkerFixedBlockSizeFlag)
		fillSize = *chunkerFixedBlockSizeFlag
		log.Println("Using fixed block size chunker")
	}

	chunker := ch.NewFillDetectingChunker(baseChunker, fillSize-100)

	opts := badger.DefaultOptions
	opts.Dir = *dbDirectory
	opts.ValueDir = *dbDirectory

	db, err := badger.Open(opts)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	traversePipe := internal.NewTraversePipeline(pathChannel)
	chunkPipe := internal.NewChunkPipeline(traversePipe.Output(), chunker)
	hashPipe := internal.NewHashPipeline(chunkPipe.Output(), hasher)
	storePipe := internal.NewStorePipeline(hashPipe.Output(), db, 4)

	go traversePipe.Process()
	go chunkPipe.Process()
	go hashPipe.Process()
	go storePipe.Process()

	log.Println("Sending path to process")

	pathChannel <- *dataDirectory
	close(pathChannel)

	log.Println("Waiting for finish")

	<-storePipe.Output()

	log.Println("Done")

	spew.Dump(traversePipe.Metrics(), chunkPipe.Metrics(), hashPipe.Metrics(), storePipe.Metrics())
}
