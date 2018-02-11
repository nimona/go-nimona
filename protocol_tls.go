package fabric

import (
	"context"
	"crypto/tls"
)

// SecProtocol is a TLS protocol
type SecProtocol struct {
	Config tls.Config
}

// Name of the protocol
func (m *SecProtocol) Name() string {
	return "tls"
}

// Handle is the protocol handler for the server
func (m *SecProtocol) Handle(ctx context.Context, c Conn) (context.Context, Conn, error) {
	scon := tls.Server(c, &m.Config)
	if err := scon.Handshake(); err != nil {
		return nil, nil, err
	}

	nc := newConnWrapper(scon, c.GetAddress())
	return ctx, nc, nil
}

// Negotiate handles the client's side of the tls protocol
func (m *SecProtocol) Negotiate(ctx context.Context, c Conn) (context.Context, Conn, error) {
	scon := tls.Client(c, &m.Config)
	if err := scon.Handshake(); err != nil {
		return ctx, nil, err
	}

	nc := newConnWrapper(scon, c.GetAddress())
	return ctx, nc, nil
}
