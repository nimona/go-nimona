package protocol

import (
	"context"

	conn "github.com/nimona/go-nimona-fabric/connection"
)

// Protocol deals with both sides, server and client
type Protocol interface {
	Handle(HandlerFunc) HandlerFunc
	Negotiate(NegotiatorFunc) NegotiatorFunc
	Name() string
}

// HandlerFunc for protocol.Handle
type HandlerFunc func(ctx context.Context, c conn.Conn) error

// NegotiatorFunc for protocol.Negotiate
type NegotiatorFunc func(ctx context.Context, c conn.Conn) error

// HandlerChain -
func HandlerChain(fns ...Protocol) HandlerFunc {
	if len(fns) == 0 {
		return nil
	}
	return fns[0].Handle(HandlerChain(fns[1:cap(fns)]...))
}

// NegotiatorChain -
func NegotiatorChain(fns ...Protocol) NegotiatorFunc {
	if len(fns) == 0 {
		return nil
	}
	return fns[0].Negotiate(NegotiatorChain(fns[1:cap(fns)]...))
}
