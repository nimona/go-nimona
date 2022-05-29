package provider

import (
	"testing"

	"nimona.io/internal/connmanager"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/peer"

	"github.com/stretchr/testify/require"
)

func NewTestProvider(
	ctx context.Context,
	t *testing.T,
) (*Provider, *peer.ConnectionInfo) {
	// construct new key
	key, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	// construct new network
	con := connmanager.New(key)

	// start listening
	lis, err := con.Listen(
		ctx,
		"0.0.0.0:0",
		&connmanager.ListenConfig{
			BindLocal:   true,
			BindPrivate: true,
		},
	)
	require.NoError(t, err)

	// close on test end
	t.Cleanup(func() {
		lis.Close()
	})

	// construct new hyperspace provider
	p, err := New(
		ctx,
		con,
		key,
		nil,
	)
	require.NoError(t, err)

	// return provider and connection info
	return p, &peer.ConnectionInfo{
		Owner:     peer.IDFromPublicKey(key.PublicKey()),
		Addresses: con.Addresses(),
	}
}
