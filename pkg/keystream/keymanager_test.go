package keystream

import (
	"database/sql"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/configstore"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/networkmock"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
)

func TestKeyManager(t *testing.T) {
	sqlStoreDB, err := sql.Open(
		"sqlite",
		path.Join(t.TempDir(), "object.sqlite"),
	)
	require.NoError(t, err)

	sqlStore, err := sqlobjectstore.New(sqlStoreDB)
	require.NoError(t, err)
	configStoreDB, err := sql.Open(
		"sqlite",
		path.Join(t.TempDir(), "config.sqlite"),
	)
	require.NoError(t, err)

	configStore, err := configstore.NewSQLProvider(configStoreDB)
	require.NoError(t, err)

	k1, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	net := &networkmock.MockNetworkSimple{
		ReturnConnectionInfo: &peer.ConnectionInfo{
			Owner: peer.IDFromPublicKey(k1.PublicKey()),
		},
	}

	// construct a new manager
	sm1, err := stream.NewManager(
		context.New(),
		nil,
		nil,
		sqlStore,
	)
	require.NoError(t, err)
	m1, err := NewKeyManager(
		net,
		sqlStore,
		sm1,
		configStore,
	)
	require.NoError(t, err)

	// add a new controller
	c1, err := m1.NewController(nil)
	require.NoError(t, err)

	// check we set controller
	gc, err := m1.GetController()
	require.Equal(t, c1, gc)
	require.NoError(t, err)

	// construct a new manager that should restore the previous controller
	sm2, err := stream.NewManager(
		context.New(),
		nil,
		nil,
		sqlStore,
	)
	require.NoError(t, err)
	m2, err := NewKeyManager(
		net,
		sqlStore,
		sm2,
		configStore,
	)
	require.NoError(t, err)

	// check that controller exists
	gc, err = m2.GetController()
	require.NoError(t, err)
	require.Equal(t, c1.GetKeyStream().Root, gc.GetKeyStream().Root)
	require.Equal(t, c1.CurrentKey(), gc.CurrentKey())
	require.Equal(t, c1.GetKeyStream().Sequence, gc.GetKeyStream().Sequence)
}
