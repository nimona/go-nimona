package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func NewTestNodeConfig(t *testing.T) *NodeConfig {
	t.Helper()

	pub, prv, err := GenerateKey()
	require.NoError(t, err)

	transport := &TransportUTP{}
	listener, err := transport.Listen(context.Background(), "127.0.0.1:0")
	require.NoError(t, err)

	return &NodeConfig{
		Dialer:   transport,
		Listener: listener,
		PeerConfig: NewPeerConfig(
			prv,
			pub,
			&PeerInfo{
				PublicKey: pub,
				Addresses: []PeerAddr{
					listener.PeerAddr(),
				},
			},
		),
		DocumentStore: NewTestDocumentStore(t),
	}
}
