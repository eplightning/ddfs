package chunker

import (
	"io"

	restic "github.com/restic/chunker"
)

// RabinChunker is a chunker using variable block size
type RabinChunker struct {
	*restic.Chunker
	poly    restic.Pol
	minSize uint
	maxSize uint
}

// NewRabinChunker creates new variable chunker
func NewRabinChunker(minSize, maxSize int, poly uint64) *RabinChunker {
	polynomial := restic.Pol(poly)

	return &RabinChunker{
		Chunker: restic.NewWithBoundaries(nil, polynomial, uint(minSize), uint(maxSize)),
		poly:    polynomial,
		maxSize: uint(maxSize),
		minSize: uint(minSize),
	}
}

// NextChunk reads next chunk
func (chunker *RabinChunker) NextChunk() (Chunk, error) {
	buffer := make([]byte, chunker.maxSize)
	chunk, err := chunker.Chunker.Next(buffer)

	if err == nil {
		return NewRawChunk(chunk.Data), nil
	}

	return nil, err
}

// Reset resets reader instance
func (chunker *RabinChunker) Reset(reader io.Reader) {
	chunker.Chunker.ResetWithBoundaries(reader, chunker.poly, chunker.minSize, chunker.maxSize)
}
