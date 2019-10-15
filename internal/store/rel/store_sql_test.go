package rel_test

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"nimona.io/internal/store/rel"
	"nimona.io/pkg/hash"
	"nimona.io/pkg/stream"
)

const dbFilepath string = "./nimona.db"

func TestNewDatabase(t *testing.T) {
	dblite, err := sql.Open("sqlite3", dbFilepath)
	defer func() {
		os.Remove(dbFilepath)
	}()
	require.NoError(t, err)

	db, err := rel.New(dblite)
	require.NoError(t, err)
	require.NotNil(t, db)

	err = db.Close()
	require.NoError(t, err)
}

func TestStoreRetrieveUpdate(t *testing.T) {
	dblite, err := sql.Open("sqlite3", dbFilepath)
	defer func() {
		os.Remove(dbFilepath)
	}()
	require.NoError(t, err)

	db, err := rel.New(dblite)
	require.NoError(t, err)
	require.NotNil(t, db)

	p := stream.Created{
		Nonce: "asdf",
	}
	c := stream.PolicyAttached{
		Stream: hash.New(p.ToObject()),
	}
	obj := c.ToObject()
	obj.Set("key:s", "value")

	err = db.Store(
		obj,
		rel.WithTTL(0),
	)
	require.NoError(t, err)

	retrievedObj, err := db.GetByHash(*hash.New(obj))
	require.NoError(t, err)

	val := retrievedObj.Get("key:s")
	require.NotNil(t, val)
	assert.Equal(t, "value", val.(string))

	stHash := stream.Stream(obj)
	require.NotEmpty(t, stHash)

	err = db.UpdateTTL(*hash.New(obj), 10)
	require.NoError(t, err)

	err = db.Close()
	require.NoError(t, err)
}
