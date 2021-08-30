package keystream

import (
	"database/sql"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/networkmock"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
)

func TestKeyManager(t *testing.T) {
	sqlStoreDB, err := sql.Open(
		"sqlite3",
		path.Join(t.TempDir(), "db.sqlite"),
	)
	require.NoError(t, err)
	sqlStore, err := sqlobjectstore.New(sqlStoreDB)
	require.NoError(t, err)

	k1, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	net := &networkmock.MockNetworkSimple{
		ReturnConnectionInfo: &peer.ConnectionInfo{
			Metadata: object.Metadata{
				Owner: k1.PublicKey().DID(),
			},
		},
	}

	// construct a new manager
	m1, err := NewKeyManager(
		net,
		sqlStore,
	)
	require.NoError(t, err)

	// add a new controller
	c1, err := m1.NewController(nil)
	require.NoError(t, err)

	// check that controller exists
	ms := m1.ListControllers()
	require.Equal(t, 1, len(ms))
	require.Equal(t, c1, ms[0])

	// construct a new manager that should restore the previous controller
	m2, err := NewKeyManager(
		net,
		sqlStore,
	)
	require.NoError(t, err)

	// check that controller exists
	ms = m2.ListControllers()
	require.Equal(t, 1, len(ms))
	require.Equal(t, c1.GetKeyStream().Root, ms[0].GetKeyStream().Root)
	require.Equal(t, c1.CurrentKey(), ms[0].CurrentKey())
	require.Equal(t, c1.GetKeyStream().Sequence, ms[0].GetKeyStream().Sequence)
}
