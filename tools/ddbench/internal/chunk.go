package internal

import "git.eplight.org/eplightning/ddfs/pkg/chunker"

type ChunkToProcess struct {
	Chunk []byte
}

type ChunkMetrics struct {
	Chunks int64
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
		var chunk []byte

		for err == nil {
			chunk, err = pipeline.chunker.NextChunk()

			if len(chunk) > 0 {
				pipeline.metrics.Chunks++

				pipeline.output <- ChunkToProcess{
					Chunk: chunk,
				}
			}
		}

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
