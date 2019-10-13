package rel_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"nimona.io/internal/store/rel"
)

const dbFilepath string = "./nimona.db"

func TestNewDatabase(t *testing.T) {
	dblite, err := sql.Open("sqlite3", dbFilepath)
	require.NoError(t, err)

	db, err := rel.New(dblite)
	require.NoError(t, err)
	require.NotNil(t, db)

	err = db.Close()
	require.NoError(t, err)
}
