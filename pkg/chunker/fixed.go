package chunker

import "io"

// FixedChunker is a chunker using fixed block size
type FixedChunker struct {
	blockSize int
	reader    io.Reader
}

// NewFixedChunker creates new fixed chunker
func NewFixedChunker(blockSize int) *FixedChunker {
	return &FixedChunker{
		blockSize: blockSize,
		reader:    nil,
	}
}

// NextChunk reads next chunk
func (chunker *FixedChunker) NextChunk() (Chunk, error) {
	buffer := make([]byte, chunker.blockSize)
	read, err := io.ReadFull(chunker.reader, buffer)

	if err == nil || err == io.ErrUnexpectedEOF {
		return NewRawChunk(buffer[0:read]), nil
	}

	return nil, err
}

// Reset resets reader instance
func (chunker *FixedChunker) Reset(reader io.Reader) {
	chunker.reader = reader
}
