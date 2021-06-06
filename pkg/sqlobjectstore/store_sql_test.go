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
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/object"
	"nimona.io/pkg/object/value"
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
			Stream: object.MustMarshal(p).CID(),
		},
		Data: value.Map{
			"key": value.String("value"),
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

	fmt.Println(obj.CID())

	require.NoError(t, err)
	retrievedObj, err := store.Get(obj.CID())
	require.NoError(t, err)

	val := retrievedObj.Data["key"]
	require.NotNil(t, val)
	assert.Equal(t, "value", string(val.(value.String)))

	stCID := obj.Metadata.Stream
	require.NotEmpty(t, stCID)

	err = store.UpdateTTL(obj.CID(), 10)
	require.NoError(t, err)

	cidList, err := store.GetRelations(object.MustMarshal(p).CID())
	require.NoError(t, err)
	assert.NotEmpty(t, cidList)

	err = store.Remove(object.MustMarshal(p).CID())
	require.NoError(t, err)

	retrievedObj2, err := store.Get(object.MustMarshal(p).CID())
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

	ph := object.MustMarshal(p).CID()

	c := fixtures.TestSubscribed{}
	c.Metadata.Stream = ph

	objects := []*object.Object{
		object.MustMarshal(p),
	}

	cids := []value.CID{}
	for i := 0; i < 5; i++ {
		obj := &object.Object{
			Type: new(fixtures.TestSubscribed).Type(),
			Metadata: object.Metadata{
				Stream: ph,
				Datetime: time.Now().
					Add(time.Duration(i) * time.Hour).
					Format(time.RFC3339),
			},
			Data: value.Map{
				"keys": value.String(fmt.Sprintf("value_%d", i)),
			},
		}
		if i%2 == 0 {
			obj.Metadata.Owner = k.PublicKey()
		}
		objects = append(objects, obj)
		err = store.Put(obj)
		require.NoError(t, err)
		cids = append(cids, obj.CID())
	}

	objectReader, err := store.Filter(
		FilterByCID(cids[0]),
		FilterByCID(cids[1]),
		FilterByCID(cids[2]),
		FilterByCID(cids[3]),
		FilterByCID(cids[4]),
	)
	require.NotNil(t, objectReader)
	require.NoError(t, err)
	got, err := object.ReadAll(objectReader)
	require.NoError(t, err)
	require.Equal(t, len(cids), len(got))
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
	require.Equal(t, len(cids), len(got))

	objectReader, err = store.Filter(
		FilterByStreamCID(ph),
	)
	require.NoError(t, err)
	got, err = object.ReadAll(objectReader)
	require.NoError(t, err)
	require.Equal(t, len(cids)+1, len(got))

	t.Run("filter with limit 1 offset 0", func(t *testing.T) {
		objectReader, err = store.Filter(
			FilterByStreamCID(ph),
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
			FilterByStreamCID(ph),
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
		FilterByCID(cids[0]),
		FilterByObjectType(c.Type()),
		FilterByStreamCID(ph),
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
		Data: value.Map{
			"f00": value.String("f00"),
		},
	}

	f01 := &object.Object{
		Type: "f01",
		Metadata: object.Metadata{
			Stream: f00.CID(),
			Parents: object.Parents{
				"*": []value.CID{
					f00.CID(),
				},
			},
		},
		Data: value.Map{
			"f01": value.String("f01"),
		},
	}

	f02 := &object.Object{
		Type: "f02",
		Metadata: object.Metadata{
			Stream: f00.CID(),
			Parents: object.Parents{
				"*": []value.CID{
					f01.CID(),
				},
			},
		},
		Data: value.Map{
			"f02": value.String("f02"),
		},
	}

	fmt.Println("f00", f00.CID())
	fmt.Println("f01", f01.CID())
	fmt.Println("f02", f02.CID())

	dblite := tempSqlite3(t)
	store, err := New(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	require.NoError(t, store.Put(f00))

	t.Run("root is considered a leaf", func(t *testing.T) {
		leaves, err := store.GetStreamLeaves(f00.CID())
		require.NoError(t, err)
		require.NotNil(t, leaves)
		assert.Len(t, leaves, 1)
		assert.Equal(t, []value.CID{f00.CID()}, leaves)
	})

	require.NoError(t, store.Put(f01))
	require.NoError(t, store.Put(f02))

	leaves, err := store.GetStreamLeaves(f00.CID())
	require.NoError(t, err)
	require.NotNil(t, leaves)
	assert.Len(t, leaves, 1)
	assert.Equal(t, []value.CID{f02.CID()}, leaves)

	fmt.Println(leaves)
}

func TestStore_ListCIDs(t *testing.T) {
	f00 := &object.Object{
		Type:     "f00",
		Metadata: object.Metadata{},
		Data: value.Map{
			"f00": value.String("f00"),
		},
	}

	f01 := &object.Object{
		Type: "f01",
		Metadata: object.Metadata{
			Stream: f00.CID(),
			Parents: object.Parents{
				"*": []value.CID{
					f00.CID(),
				},
			},
		},
		Data: value.Map{
			"f01": value.String("f01"),
		},
	}

	f02 := &object.Object{
		Type: "f02",
		Data: value.Map{
			"f02": value.String("f02"),
		},
	}

	fmt.Println("f00", f00.CID())
	fmt.Println("f01", f01.CID())
	fmt.Println("f02", f02.CID())

	dblite := tempSqlite3(t)
	store, err := New(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	require.NoError(t, store.Put(f00))
	require.NoError(t, store.Put(f01))
	require.NoError(t, store.Put(f02))

	leaves, err := store.ListCIDs()
	require.NoError(t, err)
	require.NotNil(t, leaves)
	assert.Len(t, leaves, 2)
	assert.Equal(t, []value.CID{
		f00.CID(),
		f02.CID(),
	}, leaves)

	fmt.Println(leaves)
}

func TestStore_Pinned(t *testing.T) {
	dblite := tempSqlite3(t)
	store, err := New(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	t.Run("pin a", func(t *testing.T) {
		err := store.Pin("a")
		require.NoError(t, err)
	})

	t.Run("check pin a", func(t *testing.T) {
		pinned, err := store.IsPinned("a")
		require.NoError(t, err)
		assert.True(t, pinned)
	})

	t.Run("check pin x", func(t *testing.T) {
		pinned, err := store.IsPinned("x")
		require.NoError(t, err)
		assert.False(t, pinned)
	})

	t.Run("pin a again, no error", func(t *testing.T) {
		err = store.Pin("a")
		require.NoError(t, err)
	})

	t.Run("pin b", func(t *testing.T) {
		err = store.Pin("b")
		require.NoError(t, err)
	})

	t.Run("get pins (a, b)", func(t *testing.T) {
		got, err := store.GetPinned()
		require.NoError(t, err)
		require.Equal(t, []value.CID{"a", "b"}, got)
	})

	t.Run("remove pin a", func(t *testing.T) {
		err = store.RemovePin("a")
		require.NoError(t, err)
	})

	t.Run("get pins (b)", func(t *testing.T) {
		got, err := store.GetPinned()
		require.NoError(t, err)
		require.Equal(t, []value.CID{"b"}, got)
	})
}

func TestStore_GC(t *testing.T) {
	dblite := tempSqlite3(t)
	store, err := New(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	o := &object.Object{
		Type: "foo",
		Data: value.Map{
			"foo": value.String("bar"),
		},
	}

	// store object
	err = store.PutWithTTL(o, time.Second*3)
	require.NoError(t, err)

	// check object
	got, err := store.Get(o.CID())
	require.NoError(t, err)
	require.Equal(t, o, got)

	// gc()
	err = store.gc()
	require.NoError(t, err)

	// object should still be there
	got, err = store.Get(o.CID())
	require.NoError(t, err)
	require.Equal(t, o, got)

	// wait 5 seconds and check again
	time.Sleep(time.Second * 5)

	// gc()
	err = store.gc()
	require.NoError(t, err)

	// object should not be there any more
	got, err = store.Get(o.CID())
	require.Error(t, err)
	require.Nil(t, got)

	// pin object
	err = store.Pin(o.CID())
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
	got, err = store.Get(o.CID())
	require.NoError(t, err)
	require.Equal(t, o, got)
}
