package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRPCNetworkCapabilities(t *testing.T) {
	srv, clt := newTestSessionManager(t)

	caps := []string{"core/peer/capabilities"}
	hnd := &HandlerPeerCapabilities{
		Capabilities: caps,
	}
	srv.RegisterHandler(
		"core/peer/capabilities.request",
		hnd.HandlePeerCapabilitiesRequest,
	)

	// dial the server
	ses, err := clt.Dial(context.Background(), srv.NodeAddr())
	require.NoError(t, err)

	// ask for capabilities
	ctx := context.Background()
	res, err := RequestPeerCapabilities(ctx, ses)
	require.NoError(t, err)
	require.Equal(t, caps, res.Capabilities)
}
