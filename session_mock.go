package nimona

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockSession struct {
	Server *Session
	Client *Session
}

func NewMockSession(t *testing.T, skipRPC bool) *MockSession {
	t.Helper()

	sr, cw := io.Pipe()
	cr, sw := io.Pipe()

	clientConn := &MockConnEndpoint{
		Reader: sr,
		Writer: sw,
	}

	serverConn := &MockConnEndpoint{
		Reader: cr,
		Writer: cw,
	}

	m := &MockSession{
		Server: NewSession(clientConn),
		Client: NewSession(serverConn),
	}

	m.Server.skipRPC = skipRPC
	m.Client.skipRPC = skipRPC

	serverPublic, serverPrivate, err := GenerateKey()
	require.NoError(t, err)

	clientPublic, clientPrivate, err := GenerateKey()
	require.NoError(t, err)

	serverDone := make(chan struct{})

	go func() {
		err := m.Server.DoServer(serverPublic, serverPrivate)
		require.NoError(t, err)
		close(serverDone)
	}()
	err = m.Client.DoClient(clientPublic, clientPrivate)
	require.NoError(t, err)

	<-serverDone
	return m
}
