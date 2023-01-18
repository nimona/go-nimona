package nimona

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestNode_E2E(t *testing.T) {
	netHostname := NewNetworkID("testing.nimona.io")

	// Setup "server" side

	srvNodeConfig := NewTestNodeConfig(t)
	srvHandlerNetwork := &HandlerNetwork{
		Hostname: "testing.nimona.io",
		PeerAddresses: []PeerAddr{
			srvNodeConfig.Listener.PeerAddr(),
		},
	}
	srvNodeConfig.Handlers = map[string]RequestHandlerFunc{
		"core/network/info.request": srvHandlerNetwork.HandleNetworkInfoRequest,
	}
	_, err := NewNode(srvNodeConfig)
	require.NoError(t, err)

	// Setup "client" side

	clnResolver := NewMockResolver(gomock.NewController(t))
	clnResolver.EXPECT().
		Resolve(netHostname).
		Return([]PeerAddr{
			srvNodeConfig.Listener.PeerAddr(),
		}, nil)
	clnNodeConfig := NewTestNodeConfig(t)
	clnNodeConfig.Resolver = clnResolver
	clnNode, err := NewNode(clnNodeConfig)
	require.NoError(t, err)

	t.Run("join network", func(t *testing.T) {
		ctx := context.Background()
		netInfo, err := clnNode.JoinNetwork(ctx, netHostname)
		require.NoError(t, err)
		require.Equal(t, netHostname, netInfo.NetworkID)
		require.Equal(t, srvNodeConfig.Listener.PeerAddr(), netInfo.PeerAddresses[0])
	})

	t.Run("list networks", func(t *testing.T) {
		netInfos, err := clnNode.ListNetworks()
		require.NoError(t, err)
		require.Len(t, netInfos, 1)
		require.Equal(t, netHostname, netInfos[0].NetworkID)
		require.Equal(t, srvNodeConfig.Listener.PeerAddr(), netInfos[0].PeerAddresses[0])
	})
}

func NewTestNodeConfig(t *testing.T) *NodeConfig {
	t.Helper()

	pub, prv, err := GenerateKey()
	require.NoError(t, err)

	transport := &TransportUTP{}
	listener, err := transport.Listen(context.Background(), "127.0.0.1:0")
	require.NoError(t, err)

	return &NodeConfig{
		Dialer:        transport,
		Listener:      listener,
		PublicKey:     pub,
		PrivateKey:    prv,
		DocumentStore: NewTestDocumentStore(t),
	}
}
