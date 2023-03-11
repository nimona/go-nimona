package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandlerNetworkCapabilities(t *testing.T) {
	srv, clt := newTestSessionManager(t)

	caps := []string{"core/peer/capabilities"}
	HandlePeerCapabilitiesRequest(srv, caps)

	// dial the server
	ses, err := clt.Dial(context.Background(), srv.PeerAddr())
	require.NoError(t, err)

	// ask for capabilities
	ctx := context.Background()
	res, err := RequestPeerCapabilities(ctx, ses)
	require.NoError(t, err)
	require.Equal(t, caps, res.Capabilities)
}
