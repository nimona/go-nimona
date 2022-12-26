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

	expRes := &Pong{
		Nonce: "bar",
	}
	handler := func(ctx context.Context, msg *MessageRequest) error {
		fmt.Println("Server got message", msg)
		err := msg.Respond(expRes)
		require.NoError(t, err)
		return nil
	}
	srv.RegisterHandler("test/ping", handler)

	req := &Ping{
		Nonce: "foo",
	}

	// dial the server
	rpc, err := clt.Dial(context.Background(), srv.NodeAddr())
	require.NoError(t, err)

	// send a message
	gotResMsg, err := rpc.Request(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, gotResMsg)

	res := &Pong{}
	err = gotResMsg.UnmarsalInto(res)
	require.NoError(t, err)
	require.Equal(t, expRes, res)
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
