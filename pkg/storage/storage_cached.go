package storage

import "github.com/hashicorp/golang-lru"

type cachedStorage struct {
	base  Storage
	cache *lru.ARCCache
}

func NewCachedStorage(base Storage, maxEntries int) *cachedStorage {
	cache, err := lru.NewARC(maxEntries)
	if err != nil {
		panic("cannot create ARC cache")
	}

	return &cachedStorage{
		base:  base,
		cache: cache,
	}
}

func (fs *cachedStorage) Retrieve(id Identifier) ([]byte, error) {
	str := id.String()
	value, ok := fs.cache.Get(str)
	if ok {
		return value.([]byte), nil
	}

	value, err := fs.base.Retrieve(id)
	if err != nil {
		return nil, err
	}

	fs.cache.Add(str, value)
	return value.([]byte), nil
}

func (fs *cachedStorage) Remove(id Identifier) error {
	err := fs.base.Remove(id)
	fs.cache.Remove(id.String())

	return err
}

func (fs *cachedStorage) Store(id Identifier, data []byte) error {
	if err := fs.base.Store(id, data); err != nil {
		return err
	}

	fs.cache.Add(id.String(), data)
	return nil
}
