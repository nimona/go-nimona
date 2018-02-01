package fabric

import (
	"context"
)

// HandlerFunc defines the handler function for the server
type HandlerFunc func(ctx context.Context, conn Conn) (context.Context, Conn, error)

// NegotiatorFunc defines the negotiator functions for the clients
type NegotiatorFunc func(ctx context.Context, conn Conn) (context.Context, Conn, error)

// Handler is responsible for handling a negotiation on the server's side
type Handler interface {
	Handle(ctx context.Context, conn Conn) (context.Context, Conn, error)
	Name() string
}

// Negotiator is responsible for initiating a negotiation on the client's side
type Negotiator interface {
	Negotiate(ctx context.Context, conn Conn) (context.Context, Conn, error)
	Name() string
}

// Middleware are composites of a handler, a negotiator, and a name methods
type Middleware interface {
	Handle(ctx context.Context, conn Conn) (context.Context, Conn, error)
	Negotiate(ctx context.Context, conn Conn) (context.Context, Conn, error)
	Name() string
}
