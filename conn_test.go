package nimona

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConn_E2E(t *testing.T) {
	// construct new mock connection between two nodes
	mc := NewMockConn()

	// construct a new connection for the "server"
	srv := NewConn(mc.Server, 10, 10)

	// add a handler for the "server"
	go func() {
		msg, err := srv.Read()
		require.NoError(t, err)
		require.Equal(t, "ping", string(msg.Body))

		err = msg.Reply([]byte("pong"))
		require.NoError(t, err)
	}()

	// construct a new connection for the "client"
	cln := NewConn(mc.Client, 10, 10)

	// client writes to server
	res, err := cln.Request([]byte("ping"))
	require.NoError(t, err)
	require.Equal(t, "pong", string(res))
}

func TestConn_E2E_LongMessage(t *testing.T) {
	// construct new mock connection between two nodes
	mc := NewMockConn()

	// construct a new connection for the "server"
	srv := NewConn(mc.Server, 10, 10)

	// create a long message, longer than the buffer size
	body := make([]byte, 4096+100)
	for i := range body {
		body[i] = 'a'
	}

	// add a handler for the "server"
	go func() {
		msg, err := srv.Read()
		require.NoError(t, err)
		assert.Equal(t, body, msg.Body)
		msg.Reply([]byte("ok"))
	}()

	// construct a new connection for the "client"
	cln := NewConn(mc.Client, 10, 10)

	// client writes to server
	res, err := cln.Request(body)
	require.NoError(t, err)
	require.Equal(t, "ok", string(res))
}
