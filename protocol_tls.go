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
func (m *SecProtocol) Handle(fn HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c Conn) error {
		scon := tls.Server(c, &m.Config)
		if err := scon.Handshake(); err != nil {
			return err
		}

		nc := newConnWrapper(scon, c.GetAddress())
		return fn(ctx, nc)
	}
}

// Negotiate handles the client's side of the tls protocol
func (m *SecProtocol) Negotiate(fn HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c Conn) error {
		scon := tls.Client(c, &m.Config)
		if err := scon.Handshake(); err != nil {
			return err
		}

		nc := newConnWrapper(scon, c.GetAddress())
		return fn(ctx, nc)
	}
}
