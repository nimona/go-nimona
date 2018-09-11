package storage

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"nimona.io/go/blocks"
)

// diskStorage stores the block in a file
type diskStorage struct {
	path string
}

const (
	dataExt string = ".data"
	// metaExt        = ".meta"
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
	blockID, err := blocks.SumSha3(block)
	if err != nil {
		return err
	}
	dataFilePath := filepath.Join(d.path, blockID+dataExt)
	// metaFilePath := filepath.Join(d.path, block.ID()+metaExt)

	dataFileFound := false
	// metaFileFound := false

	// Check if both files exist otherwise overwrite them
	if _, err := os.Stat(dataFilePath); err == nil {
		dataFileFound = true
	}

	// if _, err := os.Stat(metaFilePath); err == nil {
	// 	metaFileFound = true
	// }

	// if dataFileFound && metaFileFound {
	if dataFileFound {
		return ErrExists
	}

	// Write the data in a file
	if err := ioutil.WriteFile(dataFilePath, block, 0644); err != nil {
		return err
	}

	// Write the meta in a file
	// mf, err := os.Create(metaFilePath)
	// if err != nil {
	// 	return err
	// }

	// defer mf.Close()

	// enc := json.NewEncoder(mf)
	// if err := enc.Encode(block); err != nil {
	// 	return err
	// }

	// mf.Sync()

	return nil
}

func (d *diskStorage) Get(key string) ([]byte, error) {
	// metaFilePath := filepath.Join(d.path, key+metaExt)
	dataFilePath := filepath.Join(d.path, key+dataExt)

	// Check if both meta and data files exist
	// _, err := os.Stat(metaFilePath)
	// if err != nil {
	// 	return nil, ErrNotFound
	// }

	if _, err := os.Stat(dataFilePath); err != nil {
		return nil, ErrNotFound
	}

	// Read bytes from the data file
	b, err := ioutil.ReadFile(dataFilePath)
	if err != nil {
		return nil, errors.New("could not read file")
	}

	return b, nil

	// Read meta from the meta file
	// mf, err := os.Open(metaFilePath)
	// if err != nil {
	// 	return nil, err
	// }

	// defer mf.Close()

	// meta := Meta{}

	// dec := json.NewDecoder(mf)
	// if err := dec.Decode(&meta); err != nil {
	// 	return nil, err
	// }

	// TODO Add block unmarhsaling
	// block, err := blocks.Unmarshal(b)
	// if err != nil {
	// 	return nil, errors.New("could not unmarshal file") // err
	// }

	// return block, nil

	// return &Block{
	// 	Key:  key,
	// 	Meta: meta,
	// 	Data: data,
	// }, nil
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
			// Check if the datafile for this key exists
			// df := filepath.Join(d.path, key+dataExt)
			// _, err = os.Stat(df)
			// if err != nil {
			// 	continue
			// }

			results = append(results, key)
		}
	}

	return results, nil
}
