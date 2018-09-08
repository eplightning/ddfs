package storage

import (
	"io/ioutil"
	"os"
	p "path"

	"github.com/pkg/errors"
)

type filesystemStorage struct {
	rootPath string
}

func NewFilesystemStorage(rootPath string) *filesystemStorage {
	return &filesystemStorage{
		rootPath: rootPath,
	}
}

func (fs *filesystemStorage) Retrieve(id Identifier) ([]byte, error) {
	str := id.String()
	path := fs.path(str)

	return ioutil.ReadFile(p.Join(path, str))
}

func (fs *filesystemStorage) Remove(id Identifier) error {
	str := id.String()
	path := fs.path(str)

	return os.Remove(p.Join(path, str))
}

func (fs *filesystemStorage) Store(id Identifier, data []byte) error {
	str := id.String()
	path := fs.path(str)

	if err := os.MkdirAll(path, 0755); err != nil {
		return errors.Wrap(err, "Could not create directories for storage")
	}

	return ioutil.WriteFile(p.Join(path, str), data, 0644)
}

func (fs *filesystemStorage) path(id string) string {
	if len(id) < 9 {
		return p.Join(fs.rootPath, "_short")
	}

	return p.Join(fs.rootPath, id[0:4], id[4:8])
}
