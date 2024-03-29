package internal

import (
	"hash"
	"time"
)

type HashToProcess struct {
	Hash  []byte
	Chunk []byte
}

type HashMetrics struct {
	Hashes      int64
	TimeElapsed time.Duration
}

type HashPipeline struct {
	input   <-chan ChunkToProcess
	output  chan HashToProcess
	metrics HashMetrics
	hasher  hash.Hash
}

func NewHashPipeline(input <-chan ChunkToProcess, hasher hash.Hash) *HashPipeline {
	return &HashPipeline{
		input:  input,
		output: make(chan HashToProcess, 1000),
		hasher: hasher,
	}
}

func (pipeline *HashPipeline) Process() {
	for chunk := range pipeline.input {
		pipeline.metrics.Hashes++

		pipeline.hasher.Reset()

		startTime := time.Now()
		pipeline.hasher.Write(chunk.Chunk)

		buffer := make([]byte, 0, 32)
		buffer = pipeline.hasher.Sum(buffer)
		pipeline.metrics.TimeElapsed = pipeline.metrics.TimeElapsed + time.Since(startTime)

		pipeline.output <- HashToProcess{
			Hash:  buffer,
			Chunk: chunk.Chunk,
		}
	}

	close(pipeline.output)
}

func (pipeline *HashPipeline) Metrics() HashMetrics {
	return pipeline.metrics
}

func (pipeline *HashPipeline) Output() <-chan HashToProcess {
	return pipeline.output
}
