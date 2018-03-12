package protocol

import (
	"context"
	"crypto/tls"
	"fmt"

	nnet "github.com/nimona/go-nimona/net"
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
func (m *SecProtocol) Handle(fn nnet.HandlerFunc) nnet.HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c nnet.Conn) error {
		scon := tls.Server(c, &m.Config)
		if err := scon.Handshake(); err != nil {
			return err
		}

		addr := c.GetAddress()
		addr.Pop()
		fmt.Println("---- TLS HANDLE AFTER POP", addr.RemainingString())

		nc := nnet.NewConnWrapper(scon, addr)
		return fn(ctx, nc)
	}
}

// Negotiate handles the client's side of the tls protocol
func (m *SecProtocol) Negotiate(fn nnet.NegotiatorFunc) nnet.NegotiatorFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c nnet.Conn) error {
		scon := tls.Client(c, &m.Config)
		if err := scon.Handshake(); err != nil {
			return err
		}

		addr := c.GetAddress()
		addr.Pop()

		nc := nnet.NewConnWrapper(scon, addr)
		return fn(ctx, nc)
	}
}
