package chunker

import (
	"io"
)

type FillChunk struct {
	Size int
	Fill byte
}

type RawChunk struct {
	Raw []byte
}

type Chunk interface {
	Bytes() []byte
}

// Chunker chunks data
type Chunker interface {
	NextChunk() (Chunk, error)
	Reset(io.Reader)
}

func NewRawChunk(raw []byte) *RawChunk {
	return &RawChunk{
		Raw: raw,
	}
}

func (chunk *RawChunk) Bytes() []byte {
	return chunk.Raw
}

func (chunk *FillChunk) Bytes() []byte {
	output := make([]byte, chunk.Size)

	for i := 0; i < chunk.Size; i++ {
		output[i] = chunk.Fill
	}

	return output
}
