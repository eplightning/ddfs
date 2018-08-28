package chunker

import "io"

type FillDetectingChunker struct {
	base        Chunker
	reader      *fillReader
	prevChunk   Chunk
	prevErr     error
	eofDetected bool
}

func NewFillDetectingChunker(base Chunker, minSize int) *FillDetectingChunker {
	return &FillDetectingChunker{
		base:   base,
		reader: newFillReader(minSize),
	}
}

func (chunker *FillDetectingChunker) Reset(reader io.Reader) {
	chunker.eofDetected = false
	chunker.prevChunk = nil
	chunker.prevErr = nil
	chunker.reader.reset(reader)
	chunker.base.Reset(chunker.reader)
}

func (chunker *FillDetectingChunker) NextChunk() (Chunk, error) {
	if chunker.reader.has() {
		return chunker.reader.pop(), nil
	}
	if chunker.prevChunk != nil || chunker.prevErr != nil {
		ch := chunker.prevChunk
		err := chunker.prevErr
		chunker.prevChunk = nil
		chunker.prevErr = nil
		return ch, err
	}

	if chunker.eofDetected {
		if chunker.reader.baseErr == nil {
			chunker.eofDetected = false
			chunker.reader.unblock()
			chunker.base.Reset(chunker.reader)
			return chunker.NextChunk()
		}

		return nil, io.EOF
	}

	chunk, err := chunker.base.NextChunk()
	if err == io.EOF {
		chunker.eofDetected = true
		err = nil
	}

	chunker.prevChunk = chunk
	chunker.prevErr = err
	return chunker.NextChunk()
}
