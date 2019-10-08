package rel_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nimona.io/internal/store/rel"
)

func TestNewDatabase(t *testing.T) {
	db, err := rel.New()
	require.NoError(t, err)
	require.NotNil(t, db)

	err = db.Close()
	require.NoError(t, err)
}
