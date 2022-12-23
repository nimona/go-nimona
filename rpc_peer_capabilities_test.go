package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRPCNetworkCapabilities(t *testing.T) {
	srv, clt := newTestConnectionManager(t)

	hnd := &HandlerPeerCapabilities{
		Capabilities: []string{"core/peer/capabilities"},
	}
	srv.RegisterHandler(
		"core/peer/capabilities.request",
		hnd.HandlePeerCapabilitiesRequest,
	)

	// dial the server
	rpc, err := clt.Dial(context.Background(), srv.NodeAddr())
	require.NoError(t, err)

	// ask for capabilities
	ctx := context.Background()
	res, err := RequestPeerCapabilities(ctx, rpc)
	require.NoError(t, err)
	require.Equal(t, []string{"core/peer/capabilities"}, res.Capabilities)
}
