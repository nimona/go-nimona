package nimona

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test for SessionManager
func TestSessionManager(t *testing.T) {
	srv, clt := newTestSessionManager(t)

	expRes := &Pong{
		Nonce: "bar",
	}
	handler := func(ctx context.Context, msg *Request) error {
		err := msg.Respond(expRes)
		require.NoError(t, err)
		return nil
	}
	srv.RegisterHandler("test/ping", handler)

	req := &Ping{
		Nonce: "foo",
	}

	// dial the server
	rpc, err := clt.Dial(context.Background(), srv.PeerAddr())
	require.NoError(t, err)

	// send a message
	msg := &Pong{}
	res, err := rpc.Request(context.Background(), req)
	require.NoError(t, err)
	err = msg.FromDocument(res.Document)
	require.NoError(t, err)
	require.Equal(t, expRes, msg)
}

func newTestSessionManager(t *testing.T) (srv *SessionManager, clt *SessionManager) {
	t.Helper()

	// create a new SessionManager for the server
	srvPub, srvPrv, err := GenerateKey()
	require.NoError(t, err)
	srvTransport := &TransportUTP{}
	srvListener, err := srvTransport.Listen(context.Background(), "127.0.0.1:0")
	require.NoError(t, err)
	srv, err = NewSessionManager(srvTransport, srvListener, srvPub, srvPrv)
	require.NoError(t, err)

	// create a new SessionManager for the client
	cltPub, cltPrv, err := GenerateKey()
	require.NoError(t, err)
	cltTransport := &TransportUTP{}
	cltListener, err := cltTransport.Listen(context.Background(), "127.0.0.1:0")
	require.NoError(t, err)
	clt, err = NewSessionManager(cltTransport, cltListener, cltPub, cltPrv)
	require.NoError(t, err)

	return srv, clt
}
