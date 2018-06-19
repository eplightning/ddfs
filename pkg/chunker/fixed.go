package chunker

import "io"

// FixedChunker is a chunker using fixed block size
type FixedChunker struct {
	blockSize int
	reader    io.Reader
	buffer    []byte
}

// NewFixedChunker creates new fixed chunker
func NewFixedChunker(blockSize int) *FixedChunker {
	return &FixedChunker{
		blockSize: blockSize,
		reader:    nil,
	}
}

// NextChunk reads next chunk
func (chunker *FixedChunker) NextChunk() ([]byte, error) {
	buffer := make([]byte, chunker.blockSize)
	read := 0

	for read < chunker.blockSize {
		readNow, err := chunker.reader.Read(buffer[read:])

		read += readNow

		if err != nil {
			return buffer[0:read], err
		}
	}

	return buffer, nil
}

// Reset resets reader instance
func (chunker *FixedChunker) Reset(reader io.Reader) {
	chunker.reader = reader
}

func (chunker *FixedChunker) Overhead() int {
	return 0
}
