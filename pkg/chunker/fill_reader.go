package chunker

import "io"

type fillReader struct {
	base       io.Reader
	detector   fillDetector
	chunks     []FillChunk
	baseErr    error
	secondRead bool
	blocked    bool
	tmpBuffer  []byte
}

func (reader *fillReader) Read(p []byte) (n int, err error) {
	if reader.baseErr != nil {
		return 0, reader.baseErr
	}
	if reader.blocked {
		return 0, io.EOF
	}

	// provided buffer needs to be at least minSize+1, make our own if provided buffer is too small
	if len(p) <= reader.detector.minSize {
		if reader.tmpBuffer == nil {
			reader.tmpBuffer = make([]byte, reader.detector.minSize+1)
		}
		n, err := reader.Read(reader.tmpBuffer)
		copy(p[0:n], reader.tmpBuffer[0:n])
		return n, err
	}

	// on first read try to find fill blocks
	if !reader.secondRead {
		n, err := reader.readAndDetect(p)
		reader.secondRead = true
		return n, err
	}

	return reader.readAndDetectSecond(p)
}

func (reader *fillReader) has() bool {
	return len(reader.chunks) > 0
}

func (reader *fillReader) pop() FillChunk {
	a := reader.chunks[0]
	reader.chunks = reader.chunks[1:]
	return a
}

func (reader *fillReader) readAndDetect(p []byte) (n int, err error) {
	offset := 0

	for {
		read, err := io.ReadFull(reader.base, p[offset:])
		read += offset
		if err == io.ErrUnexpectedEOF {
			err = io.EOF
		}
		if err != nil {
			reader.baseErr = err
		}

		consumed, more := reader.detector.process(p[0:read], err != nil)

		if consumed != read {
			offset = read - consumed
			read = offset

			if consumed > 0 {
				copy(p[0:offset], p[consumed:])
			}
		} else {
			offset = 0
			read = 0
		}

		if !more {
			if len(reader.detector.chunks) > 0 {
				reader.chunks = reader.detector.chunks
				reader.detector.chunks = make([]FillChunk, 0, 2)
			}

			return read, err
		}
	}
}

func (reader *fillReader) readAndDetectSecond(p []byte) (n int, err error) {
	read, err := io.ReadFull(reader.base, p[0:reader.detector.minSize])
	if err == io.ErrUnexpectedEOF {
		err = io.EOF
	}
	if err != nil {
		reader.baseErr = err
		return read, err
	}

	_, more := reader.detector.process(p[0:read], false)

	if !more {
		read2, err := reader.base.Read(p[read:])
		if err != nil {
			reader.baseErr = err
		}
		return read + read2, err
	}

	reader.blocked = true
	return 0, io.EOF
}

func newFillReader(minSize int) *fillReader {
	return &fillReader{
		detector: fillDetector{
			chunks:  make([]FillChunk, 0, 2),
			minSize: minSize,
		},
	}
}

func (reader *fillReader) unblock() {
	reader.secondRead = false
	reader.blocked = false
}

func (reader *fillReader) reset(base io.Reader) {
	reader.base = base
	reader.baseErr = nil
	reader.secondRead = false
	reader.blocked = false
	reader.chunks = make([]FillChunk, 0, 2)
	reader.detector.reset()
}
