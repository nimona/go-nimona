package protocol

import (
	"context"
	"crypto/tls"
	"fmt"

	conn "github.com/nimona/go-nimona-fabric/connection"
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
	return func(ctx context.Context, c conn.Conn) error {
		scon := tls.Server(c, &m.Config)
		if err := scon.Handshake(); err != nil {
			return err
		}

		addr := c.GetAddress()
		addr.Pop()
		fmt.Println("---- TLS HANDLE AFTER POP", addr.RemainingString())

		nc := conn.NewConnWrapper(scon, addr)
		return fn(ctx, nc)
	}
}

// Negotiate handles the client's side of the tls protocol
func (m *SecProtocol) Negotiate(fn NegotiatorFunc) NegotiatorFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c conn.Conn) error {
		scon := tls.Client(c, &m.Config)
		if err := scon.Handshake(); err != nil {
			return err
		}

		addr := c.GetAddress()
		addr.Pop()

		nc := conn.NewConnWrapper(scon, addr)
		return fn(ctx, nc)
	}
}
