package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace/provider"
	"nimona.io/pkg/mesh"
	"nimona.io/pkg/net"
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
	inet := net.New(peerKey)
	msh := mesh.New(
		ctx,
		inet,
		peerKey,
	)

	// start listening
	lis, err := msh.Listen(
		ctx,
		"0.0.0.0:0",
		mesh.ListenOnLocalIPs,
		mesh.ListenOnPrivateIPs,
	)
	require.NoError(t, err)

	// construct new hyperspace provider
	_, err = provider.New(
		ctx,
		inet,
		peerKey,
		nil,
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := lis.Close()
		require.NoError(t, err)
	})

	return msh.GetConnectionInfo()
}
