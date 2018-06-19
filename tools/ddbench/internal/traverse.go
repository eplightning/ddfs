package internal

import (
	"io"
	"os"
	"path/filepath"
)

type ReaderCloser interface {
	io.Reader
	io.Closer
}

type FileToProcess struct {
	Filename string
	Stream   ReaderCloser
}

type TraverseMetrics struct {
	Files int64
	Bytes int64
}

type TraversePipeline struct {
	input   <-chan string
	output  chan FileToProcess
	metrics TraverseMetrics
}

func NewTraversePipeline(input <-chan string) *TraversePipeline {
	return &TraversePipeline{
		input:  input,
		output: make(chan FileToProcess, 10),
	}
}

func (pipeline *TraversePipeline) Process() {
	for path := range pipeline.input {
		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err == nil && info.Mode().IsRegular() {
				file, err := os.Open(path)

				if err == nil {
					pipeline.metrics.Bytes += info.Size()
					pipeline.metrics.Files++

					pipeline.output <- FileToProcess{
						Filename: info.Name(),
						Stream:   file,
					}
				}
			}

			return nil
		})
	}

	close(pipeline.output)
}

func (pipeline *TraversePipeline) Metrics() TraverseMetrics {
	return pipeline.metrics
}

func (pipeline *TraversePipeline) Output() <-chan FileToProcess {
	return pipeline.output
}
