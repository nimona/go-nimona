package fabric

import (
	"context"
	"errors"
)

// EmptyProtocol allows using dynamic handler and negotiator functions
type EmptyProtocol struct {
	Handler    HandlerFunc
	Negotiator NegotiatorFunc
}

// Name of the protocol
func (m *EmptyProtocol) Name() string {
	return "empty"
}

// Handle is the protocol handler for the server
func (m *EmptyProtocol) Handle(fn HandlerFunc) HandlerFunc {
	return func(ctx context.Context, c Conn) error {
		if m.Handler == nil {
			return errors.New("Nil handler")
		}
		return m.Handler(ctx, c)
	}
}

// Negotiate handles the client's side of the identity protocol
func (m *EmptyProtocol) Negotiate(fn NegotiatorFunc) NegotiatorFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c Conn) error {
		if m.Negotiator == nil {
			return errors.New("Nil negotiator")
		}
		return m.Negotiator(ctx, c)
	}
}
