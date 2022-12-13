package nimona

import (
	"context"
	"fmt"
	"testing"

	"github.com/fxamacker/cbor"
	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/stretchr/testify/require"
)

// Test for ConnectionManager
func TestConnectionManager(t *testing.T) {
	// create a new ConnectionManager for the server
	srvPub, srvPrv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	srvTransport := &TransportUTP{}
	srvListener, err := srvTransport.Listen(context.Background(), ":0")
	require.NoError(t, err)
	srv, err := NewConnectionManager(srvTransport, srvListener, srvPub, srvPrv)
	require.NoError(t, err)

	// create a new ConnectionManager for the client
	cltPub, cltPrv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	cltTransport := &TransportUTP{}
	cltListener, err := cltTransport.Listen(context.Background(), ":0")
	require.NoError(t, err)
	clt, err := NewConnectionManager(cltTransport, cltListener, cltPub, cltPrv)
	require.NoError(t, err)

	srv.RegisterHandler("ping", func(ctx context.Context, msg *Message) error {
		fmt.Println("Server got message", msg)
		resBody := &MessageWrapper{
			Type: "pong",
		}
		resBytes, err := cbor.Marshal(resBody, cbor.EncOptions{})
		require.NoError(t, err)
		msg.Reply(resBytes)
		return nil
	})

	msg := &MessageWrapper{
		Type: "ping",
	}
	msgBytes, err := cbor.Marshal(msg, cbor.EncOptions{})
	require.NoError(t, err)

	// dial the server
	fmt.Println(">>>", srv.NodeAddr())
	conn, err := clt.Dial(context.Background(), srv.NodeAddr())
	require.NoError(t, err)

	// send a message
	res, err := conn.Request(context.Background(), msgBytes)
	require.NoError(t, err)
	fmt.Println("Client got response", res)
}
