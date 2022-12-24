package nimona

import (
	"context"
	"crypto/rand"
	"net"
	"sync"
	"testing"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/stretchr/testify/require"
)

func TestSession_E2E_Pipe(t *testing.T) {
	messagePing := []byte("ping") // client to server
	messagePong := []byte("pong") // server to client

	// Create a server and a client that are connected to each other
	server, client := net.Pipe()

	// Generate the server's static keys
	serverPublicKey, serverPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Generate the client's static keys
	clientPublicKey, clientPrivateKey, err := ed25519.GenerateKey(rand.Reader)
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

	type messagePingStruct struct {
		Test string `cbor:"test"`
	}
	type messagePongStruct struct {
		Test string `cbor:"test"`
	}
	messagePing := MessageWrapper[messagePingStruct]{ // client to server
		Type: "ping",
		Body: messagePingStruct{
			Test: "ping",
		},
	}
	messagePong := MessageWrapper[messagePongStruct]{ // server to client
		Type: "pong",
		Body: messagePongStruct{
			Test: "pong",
		},
	}

	// construct new mock connection between two nodes
	mc := NewMockSession(t, false)

	srv := mc.Server
	cln := mc.Client

	// add a handler for the "server"
	go func() {
		msg, err := srv.Read()
		require.NoError(t, err)

		reqAny := msg.Body
		req := MessageWrapper[messagePingStruct]{}
		err = req.FromAny(reqAny)
		require.NoError(t, err)
		require.EqualValues(t, messagePing, req)

		err = msg.Respond(messagePong.ToAny())
		require.NoError(t, err)
	}()

	// client writes to server
	resAny, err := cln.Request(ctx, messagePing.ToAny())
	require.NoError(t, err)
	require.NotNil(t, resAny)

	res := MessageWrapper[messagePongStruct]{}
	err = res.FromAny(*resAny)
	require.NoError(t, err)
	require.EqualValues(t, messagePong, res)

	// close the connections
	srv.Close()
	cln.Close()
}
