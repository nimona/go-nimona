package sqlobjectstore

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"

	"nimona.io/internal/fixtures"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
)

func tempSqlite3(t *testing.T) *sql.DB {
	dirPath, err := ioutil.TempDir("", "nimona-store-sql")
	require.NoError(t, err)
	fmt.Println(path.Join(dirPath, "sqlite3.db"))
	db, err := sql.Open("sqlite3", path.Join(dirPath, "sqlite3.db"))
	require.NoError(t, err)
	return db
}

func TestNewDatabase(t *testing.T) {
	dblite := tempSqlite3(t)
	store, err := New(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	err = store.Close()
	require.NoError(t, err)
}

func TestStoreRetrieveUpdate(t *testing.T) {
	dblite := tempSqlite3(t)
	store, err := New(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	p := fixtures.TestStream{
		Nonce: "asdf",
	}
	obj := &object.Object{
		Type: "foo",
		Metadata: object.Metadata{
			Stream: p.ToObject().Hash(),
		},
		Data: object.Map{
			"key": object.String("value"),
		},
	}

	err = store.Put(
		obj,
	)
	require.NoError(t, err)

	err = store.PutWithTTL(
		obj,
		10*time.Second,
	)

	fmt.Println(obj.Hash())

	require.NoError(t, err)
	retrievedObj, err := store.Get(obj.Hash())
	require.NoError(t, err)

	val := retrievedObj.Data["key"]
	require.NotNil(t, val)
	assert.Equal(t, "value", string(val.(object.String)))

	stHash := obj.Metadata.Stream
	require.NotEmpty(t, stHash)

	err = store.UpdateTTL(obj.Hash(), 10)
	require.NoError(t, err)

	hashList, err := store.GetRelations(p.ToObject().Hash())
	require.NoError(t, err)
	assert.NotEmpty(t, hashList)

	err = store.Remove(p.ToObject().Hash())
	require.NoError(t, err)

	retrievedObj2, err := store.Get(p.ToObject().Hash())
	require.True(t, errors.CausedBy(err, objectstore.ErrNotFound))
	require.Nil(t, retrievedObj2)

	err = store.Close()
	require.NoError(t, err)
}

func TestFilter(t *testing.T) {
	dblite := tempSqlite3(t)
	store, err := New(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	k, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	p := fixtures.TestStream{
		Nonce: "asdf",
	}

	sig, err := object.NewSignature(k, p.ToObject())
	require.NoError(t, err)

	p.Metadata.Signature = sig

	err = store.Put(p.ToObject())
	require.NoError(t, err)

	ph := p.ToObject().Hash()

	c := fixtures.TestSubscribed{}
	c.Metadata.Stream = ph

	objects := []*object.Object{
		p.ToObject(),
	}

	hashes := []object.Hash{}
	for i := 0; i < 5; i++ {
		obj := &object.Object{
			Type: new(fixtures.TestSubscribed).Type(),
			Metadata: object.Metadata{
				Stream: ph,
				Datetime: time.Now().
					Add(time.Duration(i) * time.Hour).
					Format(time.RFC3339),
			},
			Data: object.Map{
				"keys": object.String(fmt.Sprintf("value_%d", i)),
			},
		}
		if i%2 == 0 {
			obj.Metadata.Owner = k.PublicKey()
		}
		objects = append(objects, obj)
		err = store.Put(obj)
		require.NoError(t, err)
		hashes = append(hashes, obj.Hash())
	}

	objectReader, err := store.Filter(
		FilterByHash(hashes[0]),
		FilterByHash(hashes[1]),
		FilterByHash(hashes[2]),
		FilterByHash(hashes[3]),
		FilterByHash(hashes[4]),
	)
	require.NotNil(t, objectReader)
	require.NoError(t, err)
	got, err := object.ReadAll(objectReader)
	require.NoError(t, err)
	require.Equal(t, len(hashes), len(got))
	objectReader, err = store.Filter(
		FilterByOwner(k.PublicKey()),
	)
	require.NoError(t, err)
	got, err = object.ReadAll(objectReader)
	require.NoError(t, err)
	require.Equal(t, 3, len(got))

	objectReader, err = store.Filter(
		FilterByObjectType(c.Type()),
	)
	require.NoError(t, err)
	got, err = object.ReadAll(objectReader)
	require.NoError(t, err)
	require.Equal(t, len(hashes), len(got))

	objectReader, err = store.Filter(
		FilterByStreamHash(ph),
	)
	require.NoError(t, err)
	got, err = object.ReadAll(objectReader)
	require.NoError(t, err)
	require.Equal(t, len(hashes)+1, len(got))

	t.Run("filter with limit 1 offset 0", func(t *testing.T) {
		objectReader, err = store.Filter(
			FilterByStreamHash(ph),
			FilterLimit(1, 0),
			FilterOrderBy("MetadataDatetime"),
			FilterOrderDir("ASC"),
		)
		require.NoError(t, err)
		got, err = object.ReadAll(objectReader)
		require.NoError(t, err)
		require.Equal(t, 1, len(got))
		require.Equal(t, *objects[0], *got[0])
	})

	t.Run("filter with limit 1 offset 1", func(t *testing.T) {
		objectReader, err = store.Filter(
			FilterByStreamHash(ph),
			FilterLimit(1, 1),
			FilterOrderBy("MetadataDatetime"),
			FilterOrderDir("ASC"),
		)
		require.NoError(t, err)
		got, err = object.ReadAll(objectReader)
		require.NoError(t, err)
		require.Equal(t, 1, len(got))
		require.Equal(t, *objects[1], *got[0])
	})

	objectReader, err = store.Filter(
		FilterByHash(hashes[0]),
		FilterByObjectType(c.Type()),
		FilterByStreamHash(ph),
	)
	require.NoError(t, err)
	got, err = object.ReadAll(objectReader)
	require.NoError(t, err)
	require.Equal(t, 1, len(got))

	err = store.Close()
	require.NoError(t, err)
}

func TestStore_Relations(t *testing.T) {
	f00 := &object.Object{
		Type:     "f00",
		Metadata: object.Metadata{},
		Data: object.Map{
			"f00": object.String("f00"),
		},
	}

	f01 := &object.Object{
		Type: "f01",
		Metadata: object.Metadata{
			Stream: f00.Hash(),
			Parents: []object.Hash{
				f00.Hash(),
			},
		},
		Data: object.Map{
			"f01": object.String("f01"),
		},
	}

	f02 := &object.Object{
		Type: "f02",
		Metadata: object.Metadata{
			Stream: f00.Hash(),
			Parents: []object.Hash{
				f01.Hash(),
			},
		},
		Data: object.Map{
			"f02": object.String("f02"),
		},
	}

	fmt.Println("f00", f00.Hash())
	fmt.Println("f01", f01.Hash())
	fmt.Println("f02", f02.Hash())

	dblite := tempSqlite3(t)
	store, err := New(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	require.NoError(t, store.Put(f00))

	t.Run("root is considered a leaf", func(t *testing.T) {
		leaves, err := store.GetStreamLeaves(f00.Hash())
		require.NoError(t, err)
		require.NotNil(t, leaves)
		assert.Len(t, leaves, 1)
		assert.Equal(t, []object.Hash{f00.Hash()}, leaves)
	})

	require.NoError(t, store.Put(f01))
	require.NoError(t, store.Put(f02))

	leaves, err := store.GetStreamLeaves(f00.Hash())
	require.NoError(t, err)
	require.NotNil(t, leaves)
	assert.Len(t, leaves, 1)
	assert.Equal(t, []object.Hash{f02.Hash()}, leaves)

	fmt.Println(leaves)
}
