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
func NewRabinChunker(minSize, maxSize int) *RabinChunker {
	// randomPoly, _ := restic.RandomPolynomial()
	// log.Println(uint64(randomPoly))
	randomPoly := restic.Pol(14152035864944967)

	return &RabinChunker{
		Chunker: restic.NewWithBoundaries(nil, randomPoly, uint(minSize), uint(maxSize)),
		poly:    randomPoly,
		maxSize: uint(maxSize),
		minSize: uint(minSize),
	}
}

// NextChunk reads next chunk
func (chunker *RabinChunker) NextChunk() ([]byte, error) {
	buffer := make([]byte, chunker.maxSize)
	chunk, err := chunker.Chunker.Next(buffer)

	if err == nil {
		return chunk.Data, nil
	}

	return []byte{}, err
}

// Reset resets reader instance
func (chunker *RabinChunker) Reset(reader io.Reader) {
	chunker.Chunker.ResetWithBoundaries(reader, chunker.poly, chunker.minSize, chunker.maxSize)
}

func (chunker *RabinChunker) Overhead() int {
	return 4
}
