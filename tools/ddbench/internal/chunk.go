package internal

import (
	"time"

	"github.com/eplightning/ddfs/pkg/chunker"
)

type ChunkToProcess struct {
	Chunk []byte
}

type ChunkMetrics struct {
	Chunks      int64
	DataChunks  int64
	FillChunks  int64
	FillSize    int64
	TimeElapsed time.Duration
}

type ChunkPipeline struct {
	input   <-chan FileToProcess
	output  chan ChunkToProcess
	metrics ChunkMetrics
	chunker chunker.Chunker
}

func NewChunkPipeline(input <-chan FileToProcess, chunker chunker.Chunker) *ChunkPipeline {
	return &ChunkPipeline{
		input:   input,
		output:  make(chan ChunkToProcess, 1000),
		chunker: chunker,
	}
}

func (pipeline *ChunkPipeline) Process() {
	for file := range pipeline.input {
		pipeline.chunker.Reset(file.Stream)

		var err error
		var chunk chunker.Chunk

		func() {
			startTime := time.Now()

			for {
				chunk, err = pipeline.chunker.NextChunk()
				if err != nil || chunk == nil {
					break
				}

				pipeline.metrics.Chunks++

				switch f := chunk.(type) {
				case *chunker.FillChunk:
					pipeline.metrics.FillChunks++
					pipeline.metrics.FillSize = pipeline.metrics.FillSize + int64(f.Size)
				case *chunker.RawChunk:
					pipeline.metrics.DataChunks++
					pipeline.output <- ChunkToProcess{
						Chunk: f.Raw,
					}
				}
			}

			pipeline.metrics.TimeElapsed = pipeline.metrics.TimeElapsed + time.Since(startTime)
		}()

		file.Stream.Close()
	}

	close(pipeline.output)
}

func (pipeline *ChunkPipeline) Metrics() ChunkMetrics {
	return pipeline.metrics
}

func (pipeline *ChunkPipeline) Output() <-chan ChunkToProcess {
	return pipeline.output
}
