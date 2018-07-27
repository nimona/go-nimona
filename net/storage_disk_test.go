package net

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func cleanup(path, key string) {
	os.Remove(filepath.Join(path, key+dataExt))
}

func TestStoreBlockSuccess(t *testing.T) {
	path, _ := ioutil.TempDir("", "nimona-test-net-storage-disk")

	ds := NewDiskStorage(path)

	block := NewBlock("test", map[string]interface{}{
		"foo": "bar",
	})
	err := SetID(block)
	assert.NoError(t, err)

	err = ds.Store(block.Metadata.ID, block)
	assert.NoError(t, err)

	_, err = os.Stat(filepath.Join(path, block.Metadata.ID+dataExt))
	assert.NoError(t, err)

	cleanup(path, block.Metadata.ID)

}

func TestStoreBlockExists(t *testing.T) {
	path := os.TempDir()

	ds := NewDiskStorage(path)

	values := make(map[string][]byte)
	values["TestMetaKey"] = []byte("TestMetaValue")

	block := NewBlock("test", map[string]interface{}{
		"foo": "bar",
	})
	err := SetID(block)
	assert.NoError(t, err)

	err = ds.Store(block.Metadata.ID, block)
	assert.NoError(t, err)

	err = ds.Store(block.Metadata.ID, block)
	assert.Error(t, err)
	assert.EqualError(t, ErrExists, err.Error())

	cleanup(path, block.Metadata.ID)
}

func TestGetSuccess(t *testing.T) {
	path := os.TempDir()

	ds := NewDiskStorage(path)

	values := make(map[string][]byte)
	values["TestMetaKey"] = []byte("TestMetaValue")

	block := NewBlock("test", map[string]interface{}{
		"foo": "bar",
	})
	err := SetID(block)
	assert.NoError(t, err)

	err = ds.Store(block.Metadata.ID, block)
	assert.NoError(t, err)

	b, err := ds.Get(block.Metadata.ID)
	assert.NoError(t, err)
	assert.Equal(t, block.Metadata.ID, b.Metadata.ID)
	assert.Equal(t, block.Payload, b.Payload)
	assert.Equal(t, block.Metadata, b.Metadata)

	cleanup(path, block.Metadata.ID)
}

func TestGetFail(t *testing.T) {
	path := os.TempDir()

	ds := NewDiskStorage(path)

	key := "TestKey2"

	_, err := ds.Get(key)
	assert.Error(t, err)
	assert.EqualError(t, ErrNotFound, err.Error())
}

func TestListSuccess(t *testing.T) {
	path := os.TempDir()

	ds := NewDiskStorage(path)

	block := NewBlock("test", map[string]interface{}{
		"foo": "bar",
	})
	err := SetID(block)
	assert.NoError(t, err)

	err = ds.Store(block.Metadata.ID, block)
	assert.NoError(t, err)

	list, err := ds.List()
	assert.NoError(t, err)
	assert.Equal(t, block.Metadata.ID, list[0])

	cleanup(path, block.Metadata.ID)
}
