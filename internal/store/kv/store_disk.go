package kv

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"nimona.io/internal/errors"
)

// diskStore stores the object in a file
type diskStore struct {
	path string
}

const (
	dataExt string = ".data"
)

// NewDiskStorage creates a new diskStore struct with the given path
// the files that will be generated from this struct are stored in the path
func NewDiskStorage(path string) (Store, error) {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, err
	}

	return &diskStore{
		path: path,
	}, nil
}

// Put saves the object in two files one for the metadata and one for
// the data. The convetion used is key.meta and key.data. Returns error if
// the files cannot be created.
func (d *diskStore) Put(key string, object []byte) error {
	if strings.ContainsAny(key, "/\\") {
		return errors.New("disk store keys cannot contain / or \\")
	}

	dataFilePath := filepath.Join(d.path, key+dataExt)

	dataFileFound := false

	// Check if both files exist otherwise overwrite them
	if _, err := os.Stat(dataFilePath); err == nil {
		dataFileFound = true
	}

	if dataFileFound {
		return ErrExists
	}

	// Write the data in a file
	if err := ioutil.WriteFile(dataFilePath, object, 0644); err != nil {
		return err
	}

	return nil
}

func (d *diskStore) Get(key string) ([]byte, error) {
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

func (d *diskStore) Check(key string) error {
	dataFilePath := filepath.Join(d.path, key+dataExt)
	if _, err := os.Stat(dataFilePath); err != nil {
		return ErrNotFound
	}

	return nil
}

func (d *diskStore) Remove(key string) error {
	dataFilePath := filepath.Join(d.path, key+dataExt)
	if _, err := os.Stat(dataFilePath); err != nil {
		return ErrNotFound
	}

	// Read bytes from the data file
	err := os.Remove(dataFilePath)
	if err != nil {
		return errors.Wrap(err, errors.New("could not remove file"))
	}

	return nil
}

// List returns a list of all the object hashes that exist as files
func (d *diskStore) List() ([]string, error) {
	results := make([]string, 0, 0)

	files, err := ioutil.ReadDir(d.path)
	if err != nil {
		return nil, err
	}

	// Range over all the files in the path for objects
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

// Scan for a key prefix and return all matching keys
func (d *diskStore) Scan(prefix string) ([]string, error) {
	results := make([]string, 0, 0)

	files, err := ioutil.ReadDir(d.path)
	if err != nil {
		return nil, err
	}

	// Range over all the files in the path for objects
	for _, f := range files {
		name := f.Name()
		ext := filepath.Ext(name)

		if ext == dataExt {
			key := name[0 : len(name)-len(ext)]
			if strings.HasPrefix(key, prefix) {
				results = append(results, key)
			}
		}
	}

	return results, nil
}
