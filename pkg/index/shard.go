package index

import (
	"sync"
	"strconv"
	"git.eplight.org/eplightning/ddfs/pkg/api"
	"github.com/dgraph-io/badger"
	"github.com/hashicorp/golang-lru"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

type Shard struct {
	db       *badger.DB
	cache 	 *lru.ARCCache
	settings ShardSettings
	table    *api.IndexShard
	tableKey []byte
	slicePrefix string
	mutex sync.Mutex
}

type RangeData struct {
	Start   int64
	End     int64
	Entries []*api.IndexEntry
}

type ShardSettings struct {
	ID        string
	Size      int64
	SliceSize int32
	New       bool
}

func NewShard(settings ShardSettings, db *badger.DB, cache *lru.ARCCache) (*Shard, error) {
	shard := &Shard{
		db: db,
		cache: cache,
		settings: settings,
		table: &api.IndexShard{},
		tableKey: []byte("shards/" + settings.ID),
		slicePrefix: "slices/" + settings.ID + "/",
	}

	var err error

	if settings.New {
		err = shard.createTable()
	} else {
		err = shard.loadTable()
	}

	return shard, err
}

func (shard *Shard) GetRange(start, end int64) (*RangeData, error) {
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	ids, offLeft, offRight, start, end, err := shard.sliceIndex(start, end)
	if err != nil {
		return nil, err
	}

	var slices []*api.IndexSlice

	err = shard.db.View(func(txn *badger.Txn) error {
		slices, err = shard.loadSlices(ids, txn)
		return err
	})
	if err != nil {
		return nil, err
	}

	var entries []*api.IndexEntry
	var sliceLeft, sliceRight int64

	if len(slices) == 1 {
		entries, sliceLeft, sliceRight = shard.findEntries(slices[0], offLeft, offRight, false, false)
		end = start + sliceRight
		start += sliceLeft
	} else {
		entries, sliceLeft, sliceRight = shard.findEntries(slices[0], offLeft, -1, false, false)
		end = start + sliceRight
		start += sliceLeft

		for _, middleSlice := range slices[1:len(slices)-1] {
			middleEntries, _, sliceRight := shard.findEntries(middleSlice, 0, -1, false, false)
			entries = append(entries, middleEntries...)
			end += sliceRight
		}

		endEntries, _, sliceRight := shard.findEntries(slices[len(slices)-1], 0, offRight, false, false)
		entries = append(entries, endEntries...)
		end += sliceRight
	}

	return &RangeData{
		End: end,
		Entries: entries,
		Start: start,
	}, nil
}

func (shard *Shard) PutRange(data RangeData) error {
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	ids, offLeft, offRight, left, right, err := shard.sliceIndex(data.Start, data.End)
	if err != nil {
		return err
	}

	return shard.db.Update(func(txn *badger.Txn) error {
		return shard.putRangeTx(data, txn, ids, offLeft, offRight, left, right)
	})
}

func (shard *Shard) putRangeTx(data RangeData, txn *badger.Txn, ids []int64, offLeft, offRight, left, right int64) error {
	var entries []*api.IndexEntry
	var rightEntries []*api.IndexEntry

	slice, err := shard.slice(ids[0], txn)
	if err != nil {
		return err
	}

	if len(ids) == 1 {
		entries, _, _ = shard.findEntries(slice, offLeft, offRight, true, false)
		rightEntries, _, _ =  shard.findEntries(slice, offLeft, offRight, false, true)
	} else {
		entries, _, _ = shard.findEntries(slice, offLeft, -1, true, false)
		slice, err := shard.slice(ids[len(ids)-1], txn)
		if err != nil {
			return err
		}

		rightEntries, _, _ =  shard.findEntries(slice, 0, offRight, false, true)
	}

	entries = append(entries, data.Entries...)
	entries = append(entries, rightEntries...)

	allSlices := len(entries) / int(shard.settings.SliceSize)
	slices := make([]*api.IndexSlice, allSlices + 1)
	locations := make([]*api.IndexSliceLocation, allSlices + 1)

	currentLeft := left
	currentChunk := 0
	currentSlice := 0

	for _, entry := range entries {
		if currentChunk == int(shard.settings.SliceSize) {
			currentChunk = 0
			locations[currentSlice].End = currentLeft
			currentSlice++
		}
		if currentChunk == 0 {
			capacity := int(shard.settings.SliceSize)
			if allSlices == currentSlice {
				capacity = len(entries) % int(shard.settings.SliceSize)
			}

			slices[currentSlice] = &api.IndexSlice{
				Entries: make([]*api.IndexEntry, 0, capacity),
			}
			locations[currentSlice] = &api.IndexSliceLocation{
				Id: shard.table.SliceCounter,
				Start: currentLeft,
			}
			shard.table.SliceCounter++
		}

		slices[currentSlice].Entries = append(slices[currentSlice].Entries, entry)
		currentChunk++
		currentLeft += entry.BlockSize
	}
	locations[currentSlice].End = currentLeft

	for i, location := range locations {
		err := shard.updateSlice(location.Id, slices[i], txn)
		if err != nil {
			return err
		}
	}

	newIndex := make([]*api.IndexSliceLocation, 0, len(shard.table.Slices) - len(ids) + len(locations))
	copied := false

	for _, location := range shard.table.Slices {
		if location.Start < left || location.Start > right {
			newIndex = append(newIndex, location)
		} else if !copied {
			newIndex = append(newIndex, locations...)
			copied = true
		}
	}

	shard.table.Slices = newIndex
	shard.table.Revision++
	if err = shard.updateTable(txn); err != nil {
		return err
	}

	// TODO: przy GC pamiętać o tym że początek i koniec slice kopiowany jest
	// i ogólnie to do GC <SliceKey, OffStart, OffEnd>
	for _, id := range ids {
		shard.discardSlice(id, txn)
	}

	return nil
}

func (shard *Shard) sliceIndex(start, end int64) ([]int64, int64, int64, int64, int64, error) {
	if start >= end || end > shard.settings.Size {
		return nil, 0, 0, 0, 0, errors.New("invalid start and end parameters")
	}
	
	slices := make([]int64, 0, 1)

	var lastOffset, firstOffset, sliceStart, sliceEnd int64
	firstFound := false

	for _, slice := range shard.table.Slices {
		if slice.End <= start {
			continue
		}

		if !firstFound {
			firstFound = true
			firstOffset = start - slice.Start
			sliceStart = slice.Start
		}

		if slice.Start >= end {
			break
		}

		lastOffset = end - slice.Start
		sliceEnd = slice.End
		slices = append(slices, slice.Id)
	}

	if len(slices) == 0 {
		return nil, 0, 0, 0, 0, errors.New("could not find any slices")
	}
	
	return slices, firstOffset, lastOffset, sliceStart, sliceEnd, nil
}

func (shard *Shard) findEntries(slice *api.IndexSlice, start, end int64, before bool, after bool) ([]*api.IndexEntry, int64, int64) {
	entries := make([]*api.IndexEntry, 0, len(slice.Entries) / 2)

	var offset, lastOffset, firstOffset, prevOffset int64
	firstFound := false

	for _, entry := range slice.Entries {
		prevOffset = offset
		offset += entry.BlockSize

		if offset <= start {
			if before {
				entries = append(entries, entry)
			}

			continue
		}

		if before {
			break
		}

		if !after {
			if !firstFound {
				firstFound = true
				firstOffset = prevOffset
			}
			entries = append(entries, entry)
			lastOffset = offset

			if offset >= end && end != -1 {
				break
			}
		} else {
			if offset <= end {
				continue
			}

			entries = append(entries, entry)
		}
	}

	return entries, firstOffset, lastOffset
}

func (shard *Shard) loadSlices(indices []int64, txn *badger.Txn) ([]*api.IndexSlice, error) {
	slices := make([]*api.IndexSlice, len(indices))

	for i, id := range indices {
		slice, err := shard.slice(id, txn)
		if err != nil {
			return nil, err
		}
		slices[i] = slice
	}

	return slices, nil
}

func (shard *Shard) createTable() error {
	return shard.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(shard.tableKey)
		if err == nil {
			return errors.New("specified shard already exists in database")
		}
		if err != badger.ErrKeyNotFound {
			return errors.Wrap(err, "error while checking if shard table exists")
		}

		shard.table.Slices = make([]*api.IndexSliceLocation, 1)
		shard.table.SliceCounter = 1

		shard.table.Slices[0] = &api.IndexSliceLocation{
			Id: 0,
			Start: 0,
			End: shard.settings.Size,
		}

		slice := &api.IndexSlice{
			Entries: []*api.IndexEntry{shard.createFillEntry(shard.settings.Size, 0)},
		}

		err = shard.updateSlice(0, slice, txn)
		if err != nil {
			return errors.Wrap(err, "cannot create initial slice")
		}

		return errors.Wrap(shard.updateTable(txn), "error while saving new shard table")
	})
}

func (shard *Shard) loadTable() error {
	return shard.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(shard.tableKey)
		if err != nil {
			return err
		}

		value, err := item.Value()
		if err != nil {
			return err
		}

		return proto.Unmarshal(value, shard.table)
	})
}

func (shard *Shard) updateTable(txn *badger.Txn) error {
	bytes, err := proto.Marshal(shard.table)
	if err != nil {
		return errors.Wrap(err, "error while serializing shard table")
	}

	return txn.Set(shard.tableKey, bytes)
}

func (shard *Shard) updateSlice(id int64, slice *api.IndexSlice, txn *badger.Txn) error {
	bytes, err := proto.Marshal(slice)
	if err != nil {
		return errors.Wrap(err, "error while serializing slice")
	}

	key := shard.sliceKey(id)
	err = txn.Set(key, bytes)
	if err != nil {
		return errors.Wrap(err, "error while storing slice in database")
	}

	shard.cache.Add(id, slice)
	return nil
}

func (shard *Shard) discardSlice(id int64, txn *badger.Txn) error {
	key := shard.sliceKey(id)
	
	if err := txn.Delete(key); err != nil {
		return errors.Wrap(err, "error while discarding slice")
	}

	shard.cache.Remove(id)
	return nil
}

func (shard *Shard) slice(id int64, txn *badger.Txn) (*api.IndexSlice, error) {
	cached, ok := shard.cache.Get(id)
	if ok {
		return cached.(*api.IndexSlice), nil
	}

	item, err := txn.Get(shard.sliceKey(id))
	if err != nil {
		return nil, errors.Wrap(err, "Could not load slice")
	}

	value, err := item.Value()
	if err != nil {
		return nil, errors.Wrap(err, "Could not load slice value")
	}

	result := &api.IndexSlice{}
	err = proto.Unmarshal(value, result)
	if err != nil {
		return nil, errors.Wrap(err, "Could not unmarshal slice")
	}

	shard.cache.Add(id, result)
	return result, nil
}

func (shard *Shard) sliceKey(id int64) []byte {
	return []byte(shard.slicePrefix + strconv.FormatInt(id, 10))
}

func (shard *Shard) createFillEntry(size int64, b int32) *api.IndexEntry {
	return &api.IndexEntry{
		BlockSize: size,
		Entry: &api.IndexEntry_Fill{
			Fill: &api.FillIndexEntry{
				Byte: b,
			},
		},
	}
}

func (shard *Shard) createHashEntry(size int64, hash []byte) *api.IndexEntry {
	return &api.IndexEntry{
		BlockSize: size,
		Entry: &api.IndexEntry_Hash{
			Hash: &api.HashIndexEntry{
				Hash: hash,
			},
		},
	}
}
