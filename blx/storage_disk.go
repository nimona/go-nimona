package blx

import (
	"encoding/gob"
	"io/ioutil"
	"os"
	"path/filepath"
)

// diskStorage stores the block in a file
type diskStorage struct {
	path string
}

const (
	dataExt string = ".data"
	metaExt        = ".meta"
)

func newDiskStorage(path string) *diskStorage {
	return &diskStorage{
		path: path,
	}
}

// Store saves the block in two files one for the metadata and one for
// the data. The convetion used is key.meta and key.data. Returns error if
// the files cannot be created.
func (d *diskStorage) Store(key string, block *Block) error {
	dataFilePath := filepath.Join(d.path, block.Key, dataExt)
	metaFilePath := filepath.Join(d.path, block.Key, metaExt)

	// Write the data in a file
	df, err := os.Create(dataFilePath)
	if err != nil {
		return err
	}

	defer df.Close()

	if _, err := df.Write(block.Data); err != nil {
		return err
	}

	df.Sync()

	// Write the meta in a file
	mf, err := os.Create(metaFilePath)
	if err != nil {
		return err
	}

	defer mf.Close()

	// Create an encoder to store the meta map
	enc := gob.NewEncoder(mf)
	if err := enc.Encode(block.Meta); err != nil {
		return nil
	}

	mf.Sync()

	return nil
}

func (d *diskStorage) Get(key string) (*Block, error) {
	metaFilePath := filepath.Join(d.path, key, metaExt)
	dataFilePath := filepath.Join(d.path, key, dataExt)

	// Check if both meta and data files exist
	_, err := os.Stat(metaFilePath)
	if err != nil {
		return nil, ErrNotFound
	}

	_, err = os.Stat(dataFilePath)
	if err != nil {
		return nil, ErrNotFound
	}

	// Read bytes from the data file
	data, err := ioutil.ReadFile(dataFilePath)
	if err != nil {
		return nil, err
	}

	// Read meta from the meta file
	mf, err := os.Open(metaFilePath)
	if err != nil {
		return nil, err
	}

	mf.Close()

	meta := make(map[string][]byte)

	dec := gob.NewDecoder(mf)
	if err := dec.Decode(&meta); err != nil {
		return nil, err
	}

	return &Block{
		Key:  key,
		Meta: meta,
		Data: data,
	}, nil
}

func (d *diskStorage) List() ([]*string, error) {
	return nil, nil
}
