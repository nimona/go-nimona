package sqlobjectstore

import (
	"database/sql"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"

	"nimona.io/internal/fixtures"
	"nimona.io/pkg/chore"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
)

func tempSqlite3(t *testing.T) *sql.DB {
	t.Helper()
	dirPath := t.TempDir()
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
			Stream: object.MustMarshal(p).Hash(),
		},
		Data: chore.Map{
			"key": chore.String("value"),
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
	assert.Equal(t, "value", string(val.(chore.String)))

	stHash := obj.Metadata.Stream
	require.NotEmpty(t, stHash)

	err = store.UpdateTTL(obj.Hash(), 10)
	require.NoError(t, err)

	hashList, err := store.GetRelations(object.MustMarshal(p).Hash())
	require.NoError(t, err)
	assert.NotEmpty(t, hashList)

	err = store.Remove(object.MustMarshal(p).Hash())
	require.NoError(t, err)

	retrievedObj2, err := store.Get(object.MustMarshal(p).Hash())
	require.True(t, errors.Is(err, objectstore.ErrNotFound))
	require.Nil(t, retrievedObj2)

	err = store.Close()
	require.NoError(t, err)
}

func TestFilter(t *testing.T) {
	dblite := tempSqlite3(t)
	store, err := New(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	k, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	p := fixtures.TestStream{
		Nonce: "asdf",
	}

	sig, err := object.NewSignature(k, object.MustMarshal(p))
	require.NoError(t, err)

	p.Metadata.Signature = sig

	err = store.Put(object.MustMarshal(p))
	require.NoError(t, err)

	ph := object.MustMarshal(p).Hash()

	c := fixtures.TestSubscribed{}
	c.Metadata.Stream = ph

	objects := []*object.Object{
		object.MustMarshal(p),
	}

	hashes := []chore.Hash{}
	for i := 0; i < 5; i++ {
		obj := &object.Object{
			Type: new(fixtures.TestSubscribed).Type(),
			Metadata: object.Metadata{
				Stream: ph,
				Datetime: time.Now().
					Add(time.Duration(i) * time.Hour).
					Format(time.RFC3339),
			},
			Data: chore.Map{
				"keys": chore.String(fmt.Sprintf("value_%d", i)),
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
		Data: chore.Map{
			"f00": chore.String("f00"),
		},
	}

	f01 := &object.Object{
		Type: "f01",
		Metadata: object.Metadata{
			Stream: f00.Hash(),
			Parents: object.Parents{
				"*": []chore.Hash{
					f00.Hash(),
				},
			},
		},
		Data: chore.Map{
			"f01": chore.String("f01"),
		},
	}

	f02 := &object.Object{
		Type: "f02",
		Metadata: object.Metadata{
			Stream: f00.Hash(),
			Parents: object.Parents{
				"*": []chore.Hash{
					f01.Hash(),
				},
			},
		},
		Data: chore.Map{
			"f02": chore.String("f02"),
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
		assert.Equal(t, []chore.Hash{f00.Hash()}, leaves)
	})

	require.NoError(t, store.Put(f01))
	require.NoError(t, store.Put(f02))

	leaves, err := store.GetStreamLeaves(f00.Hash())
	require.NoError(t, err)
	require.NotNil(t, leaves)
	assert.Len(t, leaves, 1)
	assert.Equal(t, []chore.Hash{f02.Hash()}, leaves)

	fmt.Println(leaves)
}

func TestStore_ListHashes(t *testing.T) {
	f00 := &object.Object{
		Type:     "f00",
		Metadata: object.Metadata{},
		Data: chore.Map{
			"f00": chore.String("f00"),
		},
	}

	f01 := &object.Object{
		Type: "f01",
		Metadata: object.Metadata{
			Stream: f00.Hash(),
			Parents: object.Parents{
				"*": []chore.Hash{
					f00.Hash(),
				},
			},
		},
		Data: chore.Map{
			"f01": chore.String("f01"),
		},
	}

	f02 := &object.Object{
		Type: "f02",
		Data: chore.Map{
			"f02": chore.String("f02"),
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
	require.NoError(t, store.Put(f01))
	require.NoError(t, store.Put(f02))

	leaves, err := store.ListHashes()
	require.NoError(t, err)
	require.NotNil(t, leaves)
	assert.Len(t, leaves, 2)
	assert.Equal(t, []chore.Hash{
		f00.Hash(),
		f02.Hash(),
	}, leaves)

	fmt.Println(leaves)
}

func TestStore_Pinned(t *testing.T) {
	dblite := tempSqlite3(t)
	store, err := New(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	t.Run("pin a", func(t *testing.T) {
		err := store.Pin(chore.Hash("a"))
		require.NoError(t, err)
	})

	t.Run("check pin a", func(t *testing.T) {
		pinned, err := store.IsPinned(chore.Hash("a"))
		require.NoError(t, err)
		assert.True(t, pinned)
	})

	t.Run("check pin x", func(t *testing.T) {
		pinned, err := store.IsPinned(chore.Hash("x"))
		require.NoError(t, err)
		assert.False(t, pinned)
	})

	t.Run("pin a again, no error", func(t *testing.T) {
		err = store.Pin(chore.Hash("a"))
		require.NoError(t, err)
	})

	t.Run("pin b", func(t *testing.T) {
		err = store.Pin(chore.Hash("b"))
		require.NoError(t, err)
	})

	t.Run("get pins (a, b)", func(t *testing.T) {
		got, err := store.GetPinned()
		require.NoError(t, err)
		require.Equal(t, []chore.Hash{chore.Hash("a"), chore.Hash("b")}, got)
	})

	t.Run("remove pin a", func(t *testing.T) {
		err = store.RemovePin(chore.Hash("a"))
		require.NoError(t, err)
	})

	t.Run("get pins (b)", func(t *testing.T) {
		got, err := store.GetPinned()
		require.NoError(t, err)
		require.Equal(t, []chore.Hash{chore.Hash("b")}, got)
	})
}

func TestStore_GC(t *testing.T) {
	dblite := tempSqlite3(t)
	store, err := New(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	o := &object.Object{
		Type: "foo",
		Data: chore.Map{
			"foo": chore.String("bar"),
		},
	}

	// store object
	err = store.PutWithTTL(o, time.Second*3)
	require.NoError(t, err)

	// check object
	got, err := store.Get(o.Hash())
	require.NoError(t, err)
	require.Equal(t, o, got)

	// gc()
	err = store.gc()
	require.NoError(t, err)

	// object should still be there
	got, err = store.Get(o.Hash())
	require.NoError(t, err)
	require.Equal(t, o, got)

	// wait 5 seconds and check again
	time.Sleep(time.Second * 5)

	// gc()
	err = store.gc()
	require.NoError(t, err)

	// object should not be there any more
	got, err = store.Get(o.Hash())
	require.Error(t, err)
	require.Nil(t, got)

	// pin object
	err = store.Pin(o.Hash())
	require.NoError(t, err)

	// store object again with 1 second TTL
	err = store.PutWithTTL(o, time.Second)
	require.NoError(t, err)

	// wait 2 seconds and check again
	time.Sleep(time.Second * 2)

	// gc()
	err = store.gc()
	require.NoError(t, err)

	// object should still be there
	got, err = store.Get(o.Hash())
	require.NoError(t, err)
	require.Equal(t, o, got)
}
