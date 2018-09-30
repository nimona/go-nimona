package storage

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"nimona.io/go/primitives"
)

// diskStorage stores the block in a file
type diskStorage struct {
	path string
}

const (
	dataExt string = ".data"
)

// NewDiskStorage creates a new diskStorage struct with the given path
// the files that will be generated from this struct are stored in the path
func NewDiskStorage(path string) Storage {
	os.MkdirAll(path, os.ModePerm)

	return &diskStorage{
		path: path,
	}
}

// Store saves the block in two files one for the metadata and one for
// the data. The convetion used is key.meta and key.data. Returns error if
// the files cannot be created.
func (d *diskStorage) Store(key string, block []byte) error {
	blockID, err := primitives.SumSha3(block)
	if err != nil {
		return err
	}
	dataFilePath := filepath.Join(d.path, blockID+dataExt)

	dataFileFound := false

	// Check if both files exist otherwise overwrite them
	if _, err := os.Stat(dataFilePath); err == nil {
		dataFileFound = true
	}

	if dataFileFound {
		return ErrExists
	}

	// Write the data in a file
	if err := ioutil.WriteFile(dataFilePath, block, 0644); err != nil {
		return err
	}

	return nil
}

func (d *diskStorage) Get(key string) ([]byte, error) {
	// metaFilePath := filepath.Join(d.path, key+metaExt)
	dataFilePath := filepath.Join(d.path, key+dataExt)

	if _, err := os.Stat(dataFilePath); err != nil {
		return nil, ErrNotFound
	}

	// Read bytes from the data file
	b, err := ioutil.ReadFile(dataFilePath)
	if err != nil {
		return nil, errors.New("could not read file")
	}

	return b, nil

}

// List returns a list of all the block hashes that exist as files
func (d *diskStorage) List() ([]string, error) {
	results := make([]string, 0, 0)

	files, err := ioutil.ReadDir(d.path)
	if err != nil {
		return nil, err
	}

	// Range over all the files in the path for blocks
	for _, f := range files {
		name := f.Name()
		ext := filepath.Ext(name)

		if ext == dataExt {
			key := name[0 : len(name)-len(ext)]
			results = append(results, key)
		}
	}

	return results, nil
}
