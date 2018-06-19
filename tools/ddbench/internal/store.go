package internal

import "github.com/dgraph-io/badger"

type StoreMetrics struct {
	Found         int64
	Stored        int64
	FoundBytes    int64
	StoredBytes   int64
	MetadataBytes int64
	Errors        int64
}

type StorePipeline struct {
	input    <-chan HashToProcess
	output   chan interface{}
	metrics  StoreMetrics
	database *badger.DB
	overhead int
}

func NewStorePipeline(input <-chan HashToProcess, database *badger.DB, overhead int) *StorePipeline {
	return &StorePipeline{
		input:    input,
		output:   make(chan interface{}),
		database: database,
		overhead: overhead,
	}
}

func (pipeline *StorePipeline) Process() {
	for hash := range pipeline.input {
		pipeline.database.Update(func(txn *badger.Txn) error {
			_, err := txn.Get(hash.Hash)

			if err == badger.ErrKeyNotFound {
				err = txn.Set(hash.Hash, []byte{})

				pipeline.metrics.Stored++
				pipeline.metrics.StoredBytes += int64(len(hash.Chunk))
				pipeline.metrics.MetadataBytes += int64(pipeline.overhead) + int64(len(hash.Hash))
			} else if err == nil {
				pipeline.metrics.Found++
				pipeline.metrics.FoundBytes += int64(len(hash.Chunk))
			} else {
				pipeline.metrics.Errors++
			}

			return nil
		})
	}

	close(pipeline.output)
}

func (pipeline *StorePipeline) Metrics() StoreMetrics {
	return pipeline.metrics
}

func (pipeline *StorePipeline) Output() <-chan interface{} {
	return pipeline.output
}
