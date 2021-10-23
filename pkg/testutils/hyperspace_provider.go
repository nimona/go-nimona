package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/internal/net"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace/provider"
	"nimona.io/pkg/network"
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
	nnet := net.New(peerKey)
	net := network.New(
		ctx,
		nnet,
		peerKey,
	)

	// start listening
	lis, err := net.Listen(
		ctx,
		"0.0.0.0:0",
		network.ListenOnLocalIPs,
		network.ListenOnPrivateIPs,
	)
	require.NoError(t, err)

	// construct new hyperspace provider
	_, err = provider.New(
		ctx,
		nnet,
		peerKey,
		nil,
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := lis.Close()
		require.NoError(t, err)
	})

	return net.GetConnectionInfo()
}
