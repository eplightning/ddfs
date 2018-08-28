package main

import (
	"os"

	"git.eplight.org/eplightning/ddfs/pkg/chunker"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	file, _ := os.Open("S:\\test.txt")

	baseChunker := chunker.NewFixedChunker(128)
	fillChunker := chunker.NewFillDetectingChunker(baseChunker, 50)

	fillChunker.Reset(file)

	var err error
	var chunk chunker.Chunk

	for err == nil {
		chunk, err = fillChunker.NextChunk()

		spew.Dump(chunk)
	}
}
