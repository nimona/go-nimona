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
	srv := NewConn()
	go srv.Handle(mc.Server)

	// add a handler for the "server"
	srv.Handler = func(seq uint64, data []byte, cb func([]byte) error) error {
		require.Equal(t, "ping", string(data))
		return cb([]byte("pong"))
	}

	// construct a new connection for the "client"
	cln := NewConn()
	go cln.Handle(mc.Client)

	// client writes to server
	res, err := cln.Request([]byte("ping"))
	require.NoError(t, err)
	require.Equal(t, "pong", string(res))
}

func TestConn_E2E_LongMessage(t *testing.T) {
	// construct new mock connection between two nodes
	mc := NewMockConn()

	// construct a new connection for the "server"
	srv := NewConn()
	go srv.Handle(mc.Server)

	// create a long message, longer than the buffer size
	msg := make([]byte, 4096+100)
	for i := range msg {
		msg[i] = 'a'
	}

	// add a handler for the "server"
	srv.Handler = func(seq uint64, data []byte, cb func([]byte) error) error {
		// require.Equal(t, msg, data)
		assert.Len(t, data, len(msg))
		return cb([]byte("ok"))
	}

	// construct a new connection for the "client"
	cln := NewConn()
	go cln.Handle(mc.Client)

	// client writes to server
	res, err := cln.Request(msg)
	require.NoError(t, err)
	require.Equal(t, "ok", string(res))
}
