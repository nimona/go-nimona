package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandlerNetwork(t *testing.T) {
	srv, clt := newTestSessionManager(t)

	info := NetworkInfo{
		NetworkID: NetworkID{
			Hostname: "testing.nimona.io",
		},
		PeerAddresses: []PeerAddr{{
			Network: "utp",
			Address: "localhost:1234",
		}},
	}
	hnd := &HandlerNetwork{
		Hostname: "testing.nimona.io",
		PeerAddresses: []PeerAddr{{
			Network: "utp",
			Address: "localhost:1234",
		}},
	}
	srv.RegisterHandler(
		"core/network/info.request",
		hnd.HandleNetworkInfoRequest,
	)

	// dial the server
	ses, err := clt.Dial(context.Background(), srv.PeerAddr())
	require.NoError(t, err)

	// ask for capabilities
	ctx := context.Background()
	res, err := RequestNetworkInfo(ctx, ses)
	require.NoError(t, err)
	require.Equal(t, info.NetworkID, res.NetworkID)
	require.Equal(t, info.PeerAddresses, res.PeerAddresses)
}
