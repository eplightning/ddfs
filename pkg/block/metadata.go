package block

import (
	"git.eplight.org/eplightning/ddfs/pkg/api"
	"github.com/dgraph-io/badger"
	"github.com/hashicorp/golang-lru"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

type Identifier interface {
	String() string
}

type MetadataManager struct {
	db    *badger.DB
	cache *lru.ARCCache
}

func NewMetadataManager(db *badger.DB, maxCacheEntries int) *MetadataManager {
	cache, err := lru.NewARC(maxCacheEntries)
	if err != nil {
		panic("cannot create ARC cache")
	}

	return &MetadataManager{
		db:    db,
		cache: cache,
	}
}

func (meta *MetadataManager) Get(id Identifier) (*api.BlockMetadata, error) {
	var data *api.BlockMetadata
	var err error

	err = meta.db.View(func(txn *badger.Txn) error {
		data, err = meta.fetch(id.String(), txn)
		return err
	})
	return data, nil
}

func (meta *MetadataManager) IncReferences(id Identifier) error {
	str := id.String()

	return meta.db.Update(func(txn *badger.Txn) error {
		data, err := meta.fetch(str, txn)
		if err != nil {
			data = &api.BlockMetadata{
				References: 1,
			}
		} else {
			data.References++
		}

		return meta.replace(str, data, txn)
	})
}

func (meta *MetadataManager) DecReferences(id Identifier) error {
	str := id.String()

	return meta.db.Update(func(txn *badger.Txn) error {
		data, err := meta.fetch(str, txn)
		if err != nil {
			return nil
		}
		data.References--
		return meta.replace(str, data, txn)
	})
}

func (meta *MetadataManager) fetch(id string, txn *badger.Txn) (*api.BlockMetadata, error) {
	cached, ok := meta.cache.Get(id)
	if ok {
		return cached.(*api.BlockMetadata), nil
	}

	item, err := txn.Get(meta.key(id))
	if err != nil {
		return nil, errors.Wrap(err, "Could not load meta")
	}

	value, err := item.Value()
	if err != nil {
		return nil, errors.Wrap(err, "Could not load meta value")
	}

	result := &api.BlockMetadata{}
	err = proto.Unmarshal(value, result)
	if err != nil {
		return nil, errors.Wrap(err, "Could not unmarshal meta")
	}

	meta.cache.Add(id, result)
	return result, nil
}

func (meta *MetadataManager) replace(id string, data *api.BlockMetadata, txn *badger.Txn) error {
	bytes, err := proto.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "error while serializing metadata")
	}

	err = txn.Set(meta.key(id), bytes)
	if err != nil {
		return errors.Wrap(err, "error while storing metadata in database")
	}

	meta.cache.Add(id, data)
	return nil
}

func (meta *MetadataManager) key(id string) []byte {
	return []byte("meta/" + id)
}