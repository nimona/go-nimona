package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace/provider"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/network"
	"nimona.io/pkg/peer"
)

func NewTestBootstrapPeer(t *testing.T) *peer.ConnectionInfo {
	t.Helper()

	ctx := context.New(
		context.WithCorrelationID("test-bootstrap"),
	)

	peerKey, err := crypto.NewEd25519PrivateKey(crypto.PeerKey)
	require.NoError(t, err)

	local := localpeer.New()
	local.SetPeerKey(peerKey)

	// construct new network
	net := network.New(
		ctx,
		network.WithLocalPeer(local),
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
		net,
		nil,
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := lis.Close()
		require.NoError(t, err)
	})

	return local.GetConnectionInfo()
}
