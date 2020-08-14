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
	c := fixtures.TestSubscribed{}
	obj := c.ToObject()
	obj = obj.SetType("foo")
	obj = obj.SetStream(p.ToObject().Hash())
	obj = obj.Set("key:s", "value")

	err = store.Put(
		obj,
	)
	require.NoError(t, err)

	err = store.PutWithTimeout(
		obj,
		10*time.Second,
	)

	fmt.Println(obj.Hash())

	require.NoError(t, err)
	retrievedObj, err := store.Get(obj.Hash())
	require.NoError(t, err)

	val := retrievedObj.Raw().Value("data:m").(object.Map).Value("key:s")
	require.NotNil(t, val)
	assert.Equal(t, "value", string(val.(object.String)))

	stHash := obj.GetStream()
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
	require.True(t, retrievedObj2.IsEmpty())

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

	hashes := []object.Hash{}
	for i := 0; i < 5; i++ {
		obj := c.ToObject()
		obj = obj.SetType(c.GetType())
		obj = obj.Set("key:s", fmt.Sprintf("value_%d", i))
		if i%2 == 0 {
			obj = obj.SetOwner(k.PublicKey())
		}
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
	objects, err := object.ReadAll(objectReader)
	require.NoError(t, err)
	require.Equal(t, len(hashes), len(objects))
	objectReader, err = store.Filter(
		FilterByOwner(k.PublicKey()),
	)
	require.NoError(t, err)
	objects, err = object.ReadAll(objectReader)
	require.NoError(t, err)
	require.Equal(t, 3, len(objects))

	objectReader, err = store.Filter(
		FilterByObjectType(c.GetType()),
	)
	require.NoError(t, err)
	objects, err = object.ReadAll(objectReader)
	require.NoError(t, err)
	require.Equal(t, len(hashes), len(objects))

	objectReader, err = store.Filter(
		FilterByStreamHash(ph),
	)
	require.NoError(t, err)
	objects, err = object.ReadAll(objectReader)
	require.NoError(t, err)
	require.Equal(t, len(hashes)+1, len(objects))

	objectReader, err = store.Filter(
		FilterByHash(hashes[0]),
		FilterByObjectType(c.GetType()),
		FilterByStreamHash(ph),
	)
	require.NoError(t, err)
	objects, err = object.ReadAll(objectReader)
	require.NoError(t, err)
	require.Equal(t, 1, len(objects))

	err = store.Close()
	require.NoError(t, err)
}
