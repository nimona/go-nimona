package preferences

import (
	"database/sql"
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite"
)

func tempSqlite3(t *testing.T) *sql.DB {
	t.Helper()
	dirPath := t.TempDir()
	fmt.Println(path.Join(dirPath, "sqlite3.db"))
	db, err := sql.Open("sqlite", path.Join(dirPath, "sqlite3.db"))
	require.NoError(t, err)
	return db
}

func TestStore_Config(t *testing.T) {
	dblite := tempSqlite3(t)
	store, err := NewSQLProvider(dblite)
	require.NoError(t, err)
	require.NotNil(t, store)

	t.Run("put a=0", func(t *testing.T) {
		err := store.Put("a", "0")
		require.NoError(t, err)
	})

	t.Run("put a=1", func(t *testing.T) {
		err := store.Put("a", "1")
		require.NoError(t, err)
	})

	t.Run("put b=1", func(t *testing.T) {
		err := store.Put("b", "1")
		require.NoError(t, err)
	})

	t.Run("get config", func(t *testing.T) {
		got, err := store.Get("a")
		require.NoError(t, err)
		require.Equal(t, "1", got)
	})

	t.Run("list configs", func(t *testing.T) {
		got, err := store.List()
		require.NoError(t, err)
		require.Equal(t, map[string]string{"a": "1", "b": "1"}, got)
	})

	t.Run("del config", func(t *testing.T) {
		err := store.Remove("a")
		require.NoError(t, err)
	})

	t.Run("list configs", func(t *testing.T) {
		got, err := store.List()
		require.NoError(t, err)
		require.Equal(t, map[string]string{"b": "1"}, got)
	})
}
