package keystream

import (
	"database/sql"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
	"nimona.io/pkg/network"
	"nimona.io/pkg/sqlobjectstore"
)

func TestKeyStreamManager_Handshake(t *testing.T) {
	// construct manager and controller for delegator
	sqlStorePath0 := path.Join(t.TempDir(), "db0.sqlite")
	sqlStoreDB0, err := sql.Open("sqlite", sqlStorePath0)
	require.NoError(t, err)
	sqlStore0, err := sqlobjectstore.New(sqlStoreDB0)
	require.NoError(t, err)
	n0 := network.New(context.Background())
	l0, err := n0.Listen(
		context.Background(),
		"127.0.0.1:0",
		network.ListenOnLocalIPs,
	)
	require.NoError(t, err)
	defer l0.Close()
	m0, err := NewKeyManager(n0, sqlStore0)
	require.NoError(t, err)
	c0, err := m0.NewController(nil)
	require.NoError(t, err)

	// construct manager for initiator
	sqlStorePath1 := path.Join(t.TempDir(), "db1.sqlite")
	sqlStoreDB1, err := sql.Open("sqlite", sqlStorePath1)
	require.NoError(t, err)
	sqlStore1, err := sqlobjectstore.New(sqlStoreDB1)
	require.NoError(t, err)
	n1 := network.New(context.Background())
	l1, err := n1.Listen(
		context.Background(),
		"127.0.0.1:0",
		network.ListenOnLocalIPs,
	)
	require.NoError(t, err)
	defer l1.Close()
	m1, err := NewKeyManager(n1, sqlStore1)
	require.NoError(t, err)

	// create new delegation request
	dr, c1ch, err := m1.NewDelegationRequest(
		context.Background(), // no timeout
		Permissions{},
	)
	require.NoError(t, err)

	// pass dr to delegator handler
	err = m0.HandleDelegationRequest(
		context.Background(),
		dr,
		c0,
	)
	require.NoError(t, err)

	// and wait for the controller
	c1 := <-c1ch
	require.NotNil(t, c1)

	require.Equal(t, uint64(1), c0.GetKeyStream().Sequence)
	require.Equal(t, uint64(0), c1.GetKeyStream().Sequence)
	require.Len(t, c0.GetKeyStream().Delegates, 1)
	require.Equal(t, c0.GetKeyStream().GetDID(), c1.GetKeyStream().Delegator)
}
