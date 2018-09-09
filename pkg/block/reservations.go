package block

import (
	"sync"

	"git.eplight.org/eplightning/ddfs/pkg/api"
	"git.eplight.org/eplightning/ddfs/pkg/util"
	"github.com/dgraph-io/badger"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type ReservationManager struct {
	db           *badger.DB
	reservations *api.BlockReservations
	meta         *MetadataManager
	key          []byte
	mutex        sync.Mutex
}

func NewReservationManager(db *badger.DB, meta *MetadataManager) *ReservationManager {
	manager := &ReservationManager{
		db:   db,
		meta: meta,
		key:  []byte("reservations"),
	}
	manager.load()

	return manager
}

func (man *ReservationManager) Get(id string) (*api.BlockReservation, error) {
	man.mutex.Lock()
	defer man.mutex.Unlock()
	reservation, ok := man.reservations.Reservations[id]
	if !ok {
		return nil, errors.New("reservation not found")
	}
	return reservation, nil
}

func (man *ReservationManager) Finish(id string) error {
	man.mutex.Lock()
	defer man.mutex.Unlock()
	reservation, ok := man.reservations.Reservations[id]
	if !ok {
		return errors.New("reservation not found")
	}

	for _, hash := range reservation.Hashes {
		if err := man.meta.IncReferences(util.NewBlockHash(hash)); err != nil {
			return errors.Wrap(err, "cannot update references for reservation")
		}
	}

	delete(man.reservations.Reservations, id)
	return man.persist()
}

func (man *ReservationManager) Create(hashes [][]byte, missing []int32) (string, error) {
	reservation := &api.BlockReservation{
		Hashes:        hashes,
		MissingBlocks: missing,
	}
	id := man.newID()

	man.mutex.Lock()
	defer man.mutex.Unlock()
	man.reservations.Reservations[id] = reservation

	return id, man.persist()
}

func (man *ReservationManager) load() {
	err := man.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(man.key)
		if err != nil {
			return err
		}
		value, err := item.Value()
		if err != nil {
			return err
		}

		return proto.Unmarshal(value, man.reservations)
	})

	if err != nil {
		man.reservations = &api.BlockReservations{
			Reservations: make(map[string]*api.BlockReservation),
		}
	}
}

func (man *ReservationManager) persist() error {
	bytes, err := proto.Marshal(man.reservations)
	if err != nil {
		return errors.Wrap(err, "error while serializing reservations")
	}

	return man.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(man.key, bytes)
		if err != nil {
			return errors.Wrap(err, "error while storing reservations in database")
		}
		return nil
	})
}

func (man *ReservationManager) newID() string {
	return uuid.New().String()
}
