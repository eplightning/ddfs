package storage

type Identifier interface {
	String() string
}

type Storage interface {
	Retrieve(Identifier) ([]byte, error)
	Remove(Identifier) error
	Store(Identifier, []byte) error
}
