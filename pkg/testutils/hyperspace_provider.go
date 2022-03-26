package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/internal/connmanager"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace/provider"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func NewTestBootstrapPeer(t *testing.T) *peer.ConnectionInfo {
	t.Helper()

	ctx := context.New(
		context.WithCorrelationID("test-bootstrap"),
	)

	peerKey, err := crypto.NewEd25519PrivateKey()
	require.NoError(t, err)

	// construct new network
	con := connmanager.New(peerKey)

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

	// construct new hyperspace provider
	_, err = provider.New(
		ctx,
		con,
		peerKey,
		nil,
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := lis.Close()
		require.NoError(t, err)
	})

	return &peer.ConnectionInfo{
		Metadata: object.Metadata{
			Owner: peerKey.PublicKey().DID(),
		},
		Addresses: con.Addresses(),
	}
}
