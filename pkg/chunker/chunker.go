package chunker

import (
	"io"
)

// Chunker chunks data
type Chunker interface {
	NextChunk() ([]byte, error)
	Reset(reader io.Reader)
	Overhead() int
}
