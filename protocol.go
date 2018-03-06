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

// HandlerFunc for Handle
type HandlerFunc func(ctx context.Context, c Conn) error

// NegotiatorFunc for Negotiate
type NegotiatorFunc func(ctx context.Context, c Conn) error

// HandlerChain -
func HandlerChain(fns ...Protocol) HandlerFunc {
	if len(fns) == 0 {
		return nil
	}
	for i, fn := range fns {
		if fn != nil {
			return fn.Handle(HandlerChain(fns[i+1 : cap(fns)]...))
		}
	}
	return nil
}

// NegotiatorChain -
func NegotiatorChain(fns ...Protocol) NegotiatorFunc {
	if len(fns) == 0 {
		return nil
	}
	for i, fn := range fns {
		if fn != nil {
			return fn.Negotiate(NegotiatorChain(fns[i+1 : cap(fns)]...))
		}
	}
	return nil
}
