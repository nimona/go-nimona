package nimona

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRPCNetworkCapabilities(t *testing.T) {
	srv, clt := newTestConnectionManager(t)

	srv.RegisterHandler(
		"core/peer/capabilities.request",
		HandlePeerCapabilitiesRequest,
	)

	// dial the server
	rpc, err := clt.Dial(context.Background(), srv.NodeAddr())
	require.NoError(t, err)

	// ask for capabilities
	ctx := context.Background()
	res, err := RequestPeerCapabilities(ctx, rpc)
	require.NoError(t, err)
	require.Equal(t, []string{"core/peer/capabilities"}, res.Capabilities)
	fmt.Println("Client got response", res)
}
