package nimona

import (
	"context"
	"fmt"
	"testing"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/stretchr/testify/require"
)

// Test for SessionManager
func TestSessionManager(t *testing.T) {
	srv, clt := newTestSessionManager(t)

	res := &MessageWrapper[struct{}]{
		Type: "pong",
	}
	handler := func(ctx context.Context, msg *MessageRequest) error {
		fmt.Println("Server got message", msg)
		err := msg.Respond(res.ToAny())
		require.NoError(t, err)
		return nil
	}
	srv.RegisterHandler("ping", handler)

	msg := &MessageWrapper[any]{
		Type: "ping",
	}

	// dial the server
	rpc, err := clt.Dial(context.Background(), srv.NodeAddr())
	require.NoError(t, err)

	// send a message
	gotResAny, err := rpc.Request(context.Background(), msg.ToAny())
	require.NoError(t, err)
	require.NotNil(t, res)
	gotRes := &MessageWrapper[struct{}]{}
	err = gotRes.FromAny(*gotResAny)
	require.NoError(t, err)
	fmt.Println("Client got response", res)
}

func newTestSessionManager(t *testing.T) (srv *SessionManager, clt *SessionManager) {
	t.Helper()

	// create a new SessionManager for the server
	srvPub, srvPrv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	srvTransport := &TransportUTP{}
	srvListener, err := srvTransport.Listen(context.Background(), "127.0.0.1:0")
	require.NoError(t, err)
	srv, err = NewSessionManager(srvTransport, srvListener, srvPub, srvPrv)
	require.NoError(t, err)

	// create a new SessionManager for the client
	cltPub, cltPrv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	cltTransport := &TransportUTP{}
	cltListener, err := cltTransport.Listen(context.Background(), "127.0.0.1:0")
	require.NoError(t, err)
	clt, err = NewSessionManager(cltTransport, cltListener, cltPub, cltPrv)
	require.NoError(t, err)

	return srv, clt
}
