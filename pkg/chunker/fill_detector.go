package chunker

type fillDetector struct {
	currentByte byte
	currentSize int
	chunks      []FillChunk
	minSize     int
}

func (detector *fillDetector) process(p []byte, finish bool) (int, bool) {
	consumed := 0

	for {
		remaining := len(p)
		if remaining == 0 {
			break
		}

		if detector.currentSize == 0 {
			// if creating new block, proceed only if there's enough data in the buffer
			if remaining < detector.minSize {
				break
			}

			detector.currentByte = p[0]
		}

		found := detector.findFill(p)
		detector.currentSize += found

		// termination
		if detector.currentSize < detector.minSize {
			detector.currentSize = 0
			return consumed, false
		}

		consumed += found
		p = p[found:]

		if found != remaining {
			detector.chunks = append(detector.chunks, FillChunk{
				Fill: detector.currentByte,
				Size: detector.currentSize,
			})
			detector.currentSize = 0
		}
	}

	if finish {
		if detector.currentSize > detector.minSize {
			detector.chunks = append(detector.chunks, FillChunk{
				Fill: detector.currentByte,
				Size: detector.currentSize,
			})
		}
		detector.currentSize = 0
	}

	return consumed, !finish
}

func (detector *fillDetector) findFill(p []byte) int {
	first := detector.currentByte

	for i, b := range p {
		if b != first {
			return i
		}
	}

	return len(p)
}

func (detector *fillDetector) reset() {
	detector.chunks = make([]FillChunk, 0, 2)
	detector.currentSize = 0
}
