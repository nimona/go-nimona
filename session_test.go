package nimona

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSession_E2E_Pipe(t *testing.T) {
	messagePing := []byte("ping") // client to server
	messagePong := []byte("pong") // server to client

	// Create a server and a client that are connected to each other
	server, client := net.Pipe()

	// Generate the server's static keys
	serverPublicKey, serverPrivateKey, err := GenerateKey()
	require.NoError(t, err)

	// Generate the client's static keys
	clientPublicKey, clientPrivateKey, err := GenerateKey()
	require.NoError(t, err)

	// Perform the handshake from the server side
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		serverSession := NewSession(server)
		serverSession.skipRPC = true
		err = serverSession.DoServer(serverPublicKey, serverPrivateKey)
		require.NoError(t, err)

		// Receive the message from the server
		receivedMessage, err := serverSession.read()
		require.NoError(t, err)

		// Check that the received message is the same as the original message
		require.Equal(t, messagePing, receivedMessage)

		// Send a message from the server to the client
		_, err = serverSession.write(messagePong)
		require.NoError(t, err)

		// Done
		wg.Done()
	}()

	// Perform the handshake from the client side
	clientSession := NewSession(client)
	clientSession.skipRPC = true
	err = clientSession.DoClient(clientPublicKey, clientPrivateKey)
	require.NoError(t, err)

	// Send a message from the client to the server
	_, err = clientSession.write(messagePing)
	require.NoError(t, err)

	// Read the message from the server
	receivedMessage, err := clientSession.read()
	require.NoError(t, err)

	// Check that the received message is the same as the original message
	require.Equal(t, messagePong, receivedMessage)

	// Wait for the server to finish
	wg.Wait()
}

func TestSession_E2E_RPC(t *testing.T) {
	ctx := context.Background()

	messagePing := &Ping{ // client to server
		Nonce: "foo",
	}
	messagePong := &Pong{ // server to client
		Nonce: "bar",
	}

	// construct new mock connection between two nodes
	mc := NewMockSession(t, false)

	srv := mc.Server
	cln := mc.Client

	// add a handler for the "server"
	go func() {
		req, err := srv.Read()
		require.NoError(t, err)

		msg := Ping{}
		err = msg.FromDocumentMap(req.Document)
		require.NoError(t, err)
		require.EqualValues(t, *messagePing, msg)

		err = req.Respond(messagePong)
		require.NoError(t, err)
	}()

	// client writes to server
	msg := &Pong{}
	res, err := cln.Request(ctx, messagePing)
	require.NoError(t, err)

	err = msg.FromDocumentMap(res.Document)
	require.NoError(t, err)
	require.EqualValues(t, messagePong, msg)

	// close the connections
	srv.Close()
	cln.Close()
}
