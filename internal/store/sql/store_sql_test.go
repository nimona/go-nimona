package sql_test

import (
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	sqln "nimona.io/internal/store/sql"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/stream"
)

const dbFilepath string = "./nimona.db"

func TestNewDatabase(t *testing.T) {
	dblite, err := sql.Open("sqlite3", dbFilepath)
	defer func() {
		os.Remove(dbFilepath) // nolint
	}()
	require.NoError(t, err)

	store, err := sqln.New(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	err = store.Close()
	require.NoError(t, err)
}

func TestStoreRetrieveUpdate(t *testing.T) {
	dblite, err := sql.Open("sqlite3", dbFilepath)
	defer func() {
		os.Remove(dbFilepath) // nolint
	}()
	require.NoError(t, err)

	store, err := sqln.New(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	p := stream.Created{
		Nonce: "asdf",
	}
	c := stream.PolicyAttached{
		Stream: hash.New(p.ToObject()),
	}
	obj := c.ToObject()
	obj.Set("key:s", "value")

	err = store.Put(
		obj,
		sqln.WithTTL(0),
	)
	require.NoError(t, err)

	err = store.Put(
		obj,
		sqln.WithTTL(10),
	)

	require.NoError(t, err)
	retrievedObj, err := store.Get(hash.New(obj))
	require.NoError(t, err)

	val := retrievedObj.Get("key:s")
	require.NotNil(t, val)
	assert.Equal(t, "value", val.(string))

	stHash := stream.Stream(obj)
	require.NotEmpty(t, stHash)

	err = store.UpdateTTL(hash.New(obj), 10)
	require.NoError(t, err)

	hashList, err := store.GetRelations(hash.New(p.ToObject()))
	require.NoError(t, err)
	assert.NotEmpty(t, hashList)

	err = db.Delete(hash.New(p.ToObject()))
	require.NoError(t, err)

	retrievedObj2, err := store.Get(hash.New(p.ToObject()))
	require.True(t, errors.CausedBy(err, sqln.ErrNotFound))
	require.Nil(t, retrievedObj2)

	err = store.Close()
	require.NoError(t, err)
}

func TestSubscribe(t *testing.T) {
	// create db
	dblite, err := sql.Open("sqlite3", dbFilepath)
	defer func() {
		os.Remove(dbFilepath) // nolint
	}()
	require.NoError(t, err)

	store, err := sqln.New(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	// setup data
	p := stream.Created{
		Nonce: "asdf",
	}
	streamHash := hash.New(p.ToObject())
	c := stream.PolicyAttached{
		Stream: streamHash,
	}
	obj := c.ToObject()
	obj.Set("key:s", "value")

	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		// subscribe
		subscription, err := store.Subscribe(streamHash)
		require.NoError(t, err)

		go func() {
			hs := <-subscription.Ch
			require.NotEmpty(t, hs)
			wg.Done()
		}()
	}

	// store data
	err = store.Put(
		obj,
		sqln.WithTTL(10),
	)
	require.NoError(t, err)

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		break
	case <-time.After(1 * time.Second):
		t.Fatalf("failed to get update")
	}
}
