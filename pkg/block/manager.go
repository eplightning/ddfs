package block

import (
	"os"
	"path"

	"git.eplight.org/eplightning/ddfs/pkg/storage"
	"git.eplight.org/eplightning/ddfs/pkg/util"
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
)

type BlockManager struct {
	stor     storage.Storage
	meta     *MetadataManager
	rvm      *ReservationManager
	dataPath string
}

func NewBlockManager(dataPath string) *BlockManager {
	return &BlockManager{
		dataPath: dataPath,
	}
}

func (m *BlockManager) Get(b []byte) ([]byte, error) {
	return m.stor.Retrieve(util.NewBlockHash(b))
}

func (m *BlockManager) GetChunked(bytes [][]byte, maxChunk int) chan [][]byte {
	ch := make(chan [][]byte, 100)

	go func() {
		defer close(ch)
		var err error
		var currentSize int
		var currentChunk = make([][]byte, 0, 8)

		for _, b := range bytes {
			if currentSize >= maxChunk {
				ch <- currentChunk
				currentChunk = make([][]byte, 0, 8)
				currentSize = 0
			}
			var data []byte
			data, err = m.stor.Retrieve(util.NewBlockHash(b))
			if err != nil {
				break
			}
			currentSize += len(data)
			currentChunk = append(currentChunk, data)
		}
		if err == nil && len(currentChunk) > 0 {
			ch <- currentChunk
		}
		if err != nil {
			ch <- nil
		}
	}()

	return ch
}

func (m *BlockManager) Delete(bytes [][]byte) error {
	for _, b := range bytes {
		if err := m.meta.DecReferences(util.NewBlockHash(b)); err != nil {
			return err
		}
	}
	return nil
}

func (m *BlockManager) Reserve(bytes [][]byte) (string, []int32, error) {
	// TODO: rlock when gc is ready
	missing := make([]int32, 0, len(bytes)/2)
	for i, b := range bytes {
		res, err := m.meta.GetReferences(util.NewBlockHash(b))
		if err != nil {
			return "", nil, err
		}
		if res == 0 {
			missing = append(missing, int32(i))
		}
	}
	id, err := m.rvm.Create(bytes, missing)
	if err != nil {
		return "", nil, err
	}
	return id, missing, nil
}

func (m *BlockManager) Put(id string, input chan [][]byte) error {
	reservation, err := m.rvm.Get(id)
	if err != nil {
		return err
	}

	var missingPointer int

	for bytes := range input {
		for _, b := range bytes {
			if len(reservation.MissingBlocks) <= missingPointer {
				return errors.New("too many blocks sent")
			}
			id := reservation.MissingBlocks[missingPointer]
			hash := reservation.Hashes[id]

			err = m.stor.Store(util.NewBlockHash(hash), b)
			if err != nil {
				return err
			}

			missingPointer++
		}
	}

	if len(reservation.MissingBlocks) != missingPointer {
		return errors.New("not enough blocks sent")
	}

	return m.rvm.Finish(id)
}

func (s *BlockManager) Init() error {
	badgerPath := path.Join(s.dataPath, "db")
	if err := os.MkdirAll(badgerPath, 0755); err != nil {
		return err
	}

	opt := badger.DefaultOptions
	opt.Dir = badgerPath
	opt.ValueDir = badgerPath

	db, err := badger.Open(opt)
	if err != nil {
		return err
	}

	meta, err := NewMetadataManager(db, 1024*1024)
	if err != nil {
		return err
	}
	s.meta = meta
	s.rvm = NewReservationManager(db, meta)

	storage, err := storage.NewCachedStorage(storage.NewFilesystemStorage(path.Join(s.dataPath, "blocks")), 1024)
	if err != nil {
		return err
	}
	s.stor = storage

	return nil
}

func (s *BlockManager) Start(ctl *util.SubsystemControl) {
}
