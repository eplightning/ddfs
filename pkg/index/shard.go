package index

import (
	"time"

	"git.eplight.org/eplightning/ddfs/pkg/api"
	"github.com/dgraph-io/badger"
)

type Shard struct {
	db             *badger.DB
	settings       ShardSettings
	table          api.IndexShard
	slices         map[string]*api.IndexSlice
	accessNotifier chan<- *SliceAccessNotification
}

type Entries []*api.IndexEntry

type SliceAccessNotification struct {
	Time  time.Time
	Shard string
	Slice string
}

type ShardSettings struct {
	ID        string
	Size      int64
	SliceSize int32
	New       bool
}

func NewShard(settings ShardSettings, db *badger.DB, accessNotifier chan<- *SliceAccessNotification) (*Shard, error) {

}

func (shard *Shard) GetRange(start, end int64) (Entries, error) {

}

func (shard *Shard) PutRange(start, end int64, entries Entries) error {

}

func (shard *Shard) EvictSliceCache(slice string) error {

}

func (shard *Shard) slice(id string) (*api.IndexSlice, error) {
}
