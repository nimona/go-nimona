package provider

import (
	"testing"

	"nimona.io/internal/net"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/network"
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
	inet := net.New(key)
	nnet := network.New(
		ctx,
		inet,
		key,
	)

	// start listening
	lis, err := nnet.Listen(
		ctx,
		"0.0.0.0:0",
		network.ListenOnLocalIPs,
		network.ListenOnPrivateIPs,
	)
	require.NoError(t, err)

	// close on test end
	t.Cleanup(func() {
		lis.Close()
	})

	// construct new hyperspace provider
	p, err := New(
		ctx,
		inet,
		key,
		nil,
	)
	require.NoError(t, err)

	// return provider and connection info
	return p, nnet.GetConnectionInfo()
}
