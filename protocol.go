package fabric

import (
	"context"
)

// Protocol deals with both sides, server and client
type Protocol interface {
	Handle(HandlerFunc) HandlerFunc
	Negotiate(NegotiatorFunc) NegotiatorFunc
	Name() string
}

// HandlerFunc for protocol.Handle
type HandlerFunc func(ctx context.Context, conn Conn) error

// NegotiatorFunc for protocol.Negotiate
type NegotiatorFunc func(ctx context.Context, conn Conn) error

func handlerChain(fns ...Protocol) HandlerFunc {
	if len(fns) == 0 {
		return nil
	}
	return fns[0].Handle(handlerChain(fns[1:cap(fns)]...))
}

func negotiatorChain(fns ...Protocol) NegotiatorFunc {
	if len(fns) == 0 {
		return nil
	}
	return fns[0].Negotiate(negotiatorChain(fns[1:cap(fns)]...))
}
