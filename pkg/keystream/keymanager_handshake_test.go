package keystream

import (
	"database/sql"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/internal/net"
	"nimona.io/pkg/configstore"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/network"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/stream"
)

func TestKeyStreamManager_Handshake(t *testing.T) {
	// construct manager and controller for delegator
	sqlStorePath0 := path.Join(t.TempDir(), "object.sqlite")
	sqlStoreDB0, err := sql.Open("sqlite", sqlStorePath0)
	require.NoError(t, err)
	sqlStore0, err := sqlobjectstore.New(sqlStoreDB0)
	require.NoError(t, err)
	configStorePath0 := path.Join(t.TempDir(), "config.sqlite")
	configStoreDB0, err := sql.Open("sqlite", configStorePath0)
	require.NoError(t, err)
	configStore0, err := configstore.NewSQLProvider(configStoreDB0)
	require.NoError(t, err)
	k0, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	n0 := network.New(context.Background(), net.New(k0), k0)
	l0, err := n0.Listen(
		context.Background(),
		"127.0.0.1:0",
		network.ListenOnLocalIPs,
	)
	require.NoError(t, err)
	defer l0.Close()
	sm0, err := stream.NewManager(context.New(), nil, nil, sqlStore0)
	require.NoError(t, err)
	m0, err := NewKeyManager(n0, sqlStore0, sm0, configStore0)
	require.NoError(t, err)
	c0, err := m0.NewController(nil)
	require.NoError(t, err)

	// construct manager for initiator
	sqlStorePath1 := path.Join(t.TempDir(), "object.sqlite")
	sqlStoreDB1, err := sql.Open("sqlite", sqlStorePath1)
	require.NoError(t, err)
	sqlStore1, err := sqlobjectstore.New(sqlStoreDB1)
	require.NoError(t, err)
	configStorePath1 := path.Join(t.TempDir(), "config.sqlite")
	configStoreDB1, err := sql.Open("sqlite", configStorePath1)
	require.NoError(t, err)
	configStore1, err := configstore.NewSQLProvider(configStoreDB1)
	require.NoError(t, err)

	k1, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)
	n1 := network.New(context.Background(), net.New(k1), k1)
	l1, err := n1.Listen(
		context.Background(),
		"127.0.0.1:0",
		network.ListenOnLocalIPs,
	)
	require.NoError(t, err)
	defer l1.Close()
	sm1, err := stream.NewManager(context.New(), nil, nil, sqlStore1)
	require.NoError(t, err)
	m1, err := NewKeyManager(n1, sqlStore1, sm1, configStore1)
	require.NoError(t, err)

	// create new delegation request
	dr, c1ch, err := m1.NewDelegationRequest(
		context.Background(), // no timeout
		DelegationRequestVendor{},
		Permissions{},
	)
	require.NoError(t, err)

	// pass dr to delegator handler
	err = m0.HandleDelegationRequest(
		context.Background(),
		dr,
	)
	require.NoError(t, err)

	// and wait for the controller
	c1 := <-c1ch
	require.NotNil(t, c1)

	require.Equal(t, uint64(1), c0.GetKeyStream().Sequence)
	require.Equal(t, uint64(0), c1.GetKeyStream().Sequence)
	require.Len(t, c0.GetKeyStream().Delegates, 1)
	require.Equal(t, c0.GetKeyStream().GetDID(), c1.GetKeyStream().Delegator)

	// and check it's set corectly
	gc1, err := m1.GetController()
	require.NotNil(t, gc1)
	require.NoError(t, err)
}
