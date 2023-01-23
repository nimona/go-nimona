package nimona

import (
	"context"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestNode_Network(t *testing.T) {
	netHostname := NetworkAlias{
		Hostname: "testing.nimona.io",
	}

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
		netInfo, err := clnNode.JoinNetwork(ctx, netHostname.NetworkIdentifier())
		require.NoError(t, err)
		require.Equal(t, netHostname, netInfo.NetworkAlias)
		require.Equal(t, srvNodeConfig.Listener.PeerAddr(), netInfo.PeerAddresses[0])
	})

	t.Run("list networks", func(t *testing.T) {
		netInfos, err := clnNode.ListNetworks()
		require.NoError(t, err)
		require.Len(t, netInfos, 1)
		require.Equal(t, netHostname, netInfos[0].NetworkAlias)
		require.Equal(t, srvNodeConfig.Listener.PeerAddr(), netInfos[0].PeerAddresses[0])
	})
}
