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

// NewDiskStorage creates a new diskStorage struct with the given path
// the files that will be generated from this struct are stored in the path
func NewDiskStorage(path string) *diskStorage {
	return &diskStorage{
		path: path,
	}
}

// Store saves the block in two files one for the metadata and one for
// the data. The convetion used is key.meta and key.data. Returns error if
// the files cannot be created.
func (d *diskStorage) Store(key string, block *Block) error {
	dataFilePath := filepath.Join(d.path, block.Key+dataExt)
	metaFilePath := filepath.Join(d.path, block.Key+metaExt)

	dataFileFound := false
	metaFileFound := false

	// Check if both files exist otherwise overwrite them
	if _, err := os.Stat(dataFilePath); err == nil {
		dataFileFound = true
	}

	if _, err := os.Stat(metaFilePath); err == nil {
		metaFileFound = true
	}

	if dataFileFound && metaFileFound {
		return ErrExists
	}

	// Write the data in a file
	err := ioutil.WriteFile(dataFilePath, block.Data, 0644)
	if err != nil {
		return err
	}

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
	metaFilePath := filepath.Join(d.path, key+metaExt)
	dataFilePath := filepath.Join(d.path, key+dataExt)

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

	defer mf.Close()

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

// List returns a list of all the block hashes that exist as files
func (d *diskStorage) List() ([]*string, error) {
	results := make([]*string, 0, 0)

	files, err := ioutil.ReadDir(d.path)
	if err != nil {
		return nil, err
	}

	// Range over all the files in the path for blocks
	for _, f := range files {
		name := f.Name()
		ext := filepath.Ext(name)

		if ext == metaExt {
			key := name[0 : len(name)-len(ext)]
			// Check if the datafile for this key exists
			df := filepath.Join(d.path, key+dataExt)
			_, err = os.Stat(df)
			if err != nil {
				break
			}

			results = append(results, &key)
		}
	}

	return results, nil
}
