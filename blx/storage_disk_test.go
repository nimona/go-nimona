package blx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func cleanup(path, key string) {
	os.Remove(filepath.Join(path, key+dataExt))
	os.Remove(filepath.Join(path, key+metaExt))
}

func TestStoreBlockSuccess(t *testing.T) {
	path := os.TempDir()

	ds := newDiskStorage(path)

	meta := make(map[string][]byte)
	meta["TestMetaKey"] = []byte("TestMetaValue")

	block := Block{
		Key:  "TestKey1",
		Meta: meta,
		Data: []byte("TestData"),
	}

	err := ds.Store(block.Key, &block)
	assert.NoError(t, err)

	_, err = os.Stat(filepath.Join(path, block.Key+dataExt))
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(path, block.Key+metaExt))
	assert.NoError(t, err)

	cleanup(path, block.Key)

}

func TestStoreBlockExists(t *testing.T) {
	path := os.TempDir()

	ds := newDiskStorage(path)

	meta := make(map[string][]byte)
	meta["TestMetaKey"] = []byte("TestMetaValue")

	block := Block{
		Key:  "TestKey1",
		Meta: meta,
		Data: []byte("TestData"),
	}

	err := ds.Store(block.Key, &block)
	assert.NoError(t, err)

	err = ds.Store(block.Key, &block)
	assert.Error(t, err)
	assert.EqualError(t, ErrExists, err.Error())

	cleanup(path, block.Key)
}

func TestGetSuccess(t *testing.T) {
	path := os.TempDir()

	ds := newDiskStorage(path)

	meta := make(map[string][]byte)
	meta["TestMetaKey"] = []byte("TestMetaValue")

	block := Block{
		Key:  "TestKey1",
		Meta: meta,
		Data: []byte("TestData"),
	}

	err := ds.Store(block.Key, &block)
	assert.NoError(t, err)

	b, err := ds.Get(block.Key)
	assert.NoError(t, err)
	assert.Equal(t, block.Key, b.Key)
	assert.Equal(t, block.Data, b.Data)
	assert.Equal(t, block.Meta, b.Meta)

	cleanup(path, block.Key)
}

func TestGetFail(t *testing.T) {
	path := os.TempDir()

	ds := newDiskStorage(path)

	key := "TestKey2"

	_, err := ds.Get(key)
	assert.Error(t, err)
	assert.EqualError(t, ErrNotFound, err.Error())
}

func TestListSuccess(t *testing.T) {
	path := os.TempDir()

	ds := newDiskStorage(path)

	meta := make(map[string][]byte)
	meta["TestMetaKey"] = []byte("TestMetaValue")

	block := Block{
		Key:  "TestKey1",
		Meta: meta,
		Data: []byte("TestData"),
	}

	err := ds.Store(block.Key, &block)
	assert.NoError(t, err)

	list, err := ds.List()
	assert.NoError(t, err)
	assert.Equal(t, block.Key, *list[0])

	cleanup(path, block.Key)
}
