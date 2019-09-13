package kv

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func cleanup(path, key string) {
	os.Remove(filepath.Join(path, key+dataExt)) // nolint: errcheck
}

func TestStoreObjectSuccess(t *testing.T) {
	path, _ := ioutil.TempDir("", "nimona-test-net-storage-disk")

	ds, _ := NewDiskStorage(path)

	value := []byte("bar")
	key := "foo"

	err := ds.Put(key, value)
	assert.NoError(t, err)

	key = "foo"
	_, err = os.Stat(filepath.Join(path, key+dataExt))
	assert.NoError(t, err)

	cleanup(path, key)
}

func TestStoreObjectExists(t *testing.T) {
	path, _ := ioutil.TempDir("", "nimona-test-net-storage-disk")

	ds, _ := NewDiskStorage(path)

	values := make(map[string][]byte)
	values["TestMetaKey"] = []byte("TestMetaValue")

	value := []byte("bar")
	key := "foo"

	err := ds.Put(key, value)
	assert.NoError(t, err)

	err = ds.Put(key, value)
	assert.Error(t, err)
	assert.EqualError(t, ErrExists, err.Error())

	cleanup(path, key)
}

func TestGetSuccess(t *testing.T) {
	path, _ := ioutil.TempDir("", "nimona-test-net-storage-disk")

	ds, _ := NewDiskStorage(path)

	values := make(map[string][]byte)
	values["TestMetaKey"] = []byte("TestMetaValue")

	value := []byte("bar")
	key := "foo"

	err := ds.Put(key, value)
	assert.NoError(t, err)

	bID := "foo"
	b, err := ds.Get(bID)
	assert.NoError(t, err)
	assert.Equal(t, key, bID)
	assert.Equal(t, value, b)

	cleanup(path, key)
}

func TestGetFail(t *testing.T) {
	path, _ := ioutil.TempDir("", "nimona-test-net-storage-disk")

	ds, _ := NewDiskStorage(path)

	key := "TestKey2"

	_, err := ds.Get(key)
	assert.Error(t, err)
	assert.EqualError(t, ErrNotFound, err.Error())
}

func TestListSuccess(t *testing.T) {
	path, _ := ioutil.TempDir("", "nimona-test-net-storage-disk")

	ds, _ := NewDiskStorage(path)

	value := []byte("bar")
	key := "foo"

	err := ds.Put(key, value)
	assert.NoError(t, err)

	list, err := ds.List()
	assert.NoError(t, err)
	assert.Equal(t, key, list[0])

	cleanup(path, key)
}

func TestScanSuccess(t *testing.T) {
	path, _ := ioutil.TempDir("", "nimona-test-net-storage-disk")

	ds, _ := NewDiskStorage(path)

	pairs := map[string][]byte{
		"bar":         []byte{1},
		"foo":         []byte{1},
		"foo.bar":     []byte{1},
		"foo.foo.bar": []byte{1},
		"not-foo":     []byte{1},
		"not-foo.foo": []byte{1},
	}

	for k, v := range pairs {
		err := ds.Put(k, v)
		assert.NoError(t, err)
	}

	list, err := ds.Scan("foo")
	sort.Strings(list)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo", "foo.bar", "foo.foo.bar"}, list)

	list, err = ds.Scan("foo.")
	sort.Strings(list)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo.bar", "foo.foo.bar"}, list)

	for k := range pairs {
		cleanup(path, k)
	}
}
