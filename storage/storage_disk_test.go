package storage

// import (
// 	"io/ioutil"
// 	"os"
// 	"path/filepath"
// 	"testing"

// 	"nimona.io/go/crypto"
// 	"github.com/stretchr/testify/assert"
// )

// func cleanup(path, key string) {
// 	os.Remove(filepath.Join(path, key+dataExt))
// }

// func TestStoreBlockSuccess(t *testing.T) {
// 	path, _ := ioutil.TempDir("", "nimona-test-net-storage-disk")

// 	ds := NewDiskStorage(path)

// 	block := crypto.NewBlock("test", map[string]interface{}{
// 		"foo": "bar",
// 	})
// 	blockID, err := block.ID()
// 	assert.NoError(t, err)

// 	err = ds.Store(blockID, block)
// 	assert.NoError(t, err)

// 	blockID, err = block.ID()
// 	assert.NoError(t, err)
// 	_, err = os.Stat(filepath.Join(path, blockID+dataExt))
// 	assert.NoError(t, err)

// 	cleanup(path, blockID)
// }

// func TestStoreBlockExists(t *testing.T) {
// 	path, _ := ioutil.TempDir("", "nimona-test-net-storage-disk")

// 	ds := NewDiskStorage(path)

// 	values := make(map[string][]byte)
// 	values["TestMetaKey"] = []byte("TestMetaValue")

// 	block := crypto.NewBlock("test", map[string]interface{}{
// 		"foo": "bar",
// 	})
// 	blockID, err := block.ID()
// 	assert.NoError(t, err)

// 	blockID, err = block.ID()
// 	assert.NoError(t, err)

// 	err = ds.Store(blockID, block)
// 	assert.NoError(t, err)

// 	err = ds.Store(blockID, block)
// 	assert.Error(t, err)
// 	assert.EqualError(t, ErrExists, err.Error())

// 	cleanup(path, blockID)
// }

// func TestGetSuccess(t *testing.T) {
// 	path, _ := ioutil.TempDir("", "nimona-test-net-storage-disk")

// 	ds := NewDiskStorage(path)

// 	values := make(map[string][]byte)
// 	values["TestMetaKey"] = []byte("TestMetaValue")

// 	block := crypto.NewBlock("test", map[string]interface{}{
// 		"foo": "bar",
// 	})
// 	blockID, err := block.ID()
// 	assert.NoError(t, err)

// 	err = ds.Store(blockID, block)
// 	assert.NoError(t, err)

// 	bID, err := block.ID()
// 	assert.NoError(t, err)

// 	b, err := ds.Get(bID)
// 	assert.NoError(t, err)
// 	assert.Equal(t, blockID, bID)
// 	assert.Equal(t, block.Payload, b.Payload)
// 	assert.Equal(t, block.Metadata, b.Metadata)

// 	cleanup(path, blockID)
// }

// func TestGetFail(t *testing.T) {
// 	path, _ := ioutil.TempDir("", "nimona-test-net-storage-disk")

// 	ds := NewDiskStorage(path)

// 	key := "TestKey2"

// 	_, err := ds.Get(key)
// 	assert.Error(t, err)
// 	assert.EqualError(t, ErrNotFound, err.Error())
// }

// func TestListSuccess(t *testing.T) {
// 	path, _ := ioutil.TempDir("", "nimona-test-net-storage-disk")

// 	ds := NewDiskStorage(path)

// 	block := crypto.NewBlock("test", map[string]interface{}{
// 		"foo": "bar",
// 	})
// 	blockID, err := block.ID()
// 	assert.NoError(t, err)

// 	err = ds.Store(blockID, block)
// 	assert.NoError(t, err)

// 	list, err := ds.List()
// 	assert.NoError(t, err)
// 	blockID, err = block.ID()
// 	assert.NoError(t, err)
// 	assert.Equal(t, blockID, list[0])

// 	cleanup(path, blockID)
// }
