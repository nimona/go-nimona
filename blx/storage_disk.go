package blx

import (
	"encoding/gob"
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
	df, err := os.Create(dataFilePath)
	if err != nil {
		return err
	}

	defer df.Close()

	if _, err := df.Write(block.Data); err != nil {
		return err
	}

	df.Sync()

	metaFilePath := filepath.Join(d.path, block.Key, metaExt)
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
	return nil, nil
}

func (d *diskStorage) List() ([]*string, error) {
	return nil, nil
}
