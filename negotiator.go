package fabric

import "context"

// NegotiatorFunc defines the negotiator functions for the clients
type NegotiatorFunc func(ctx context.Context, conn Conn) (context.Context, Conn, error)

// Negotiator is responsible for initiating a negotiation on the client's side
type Negotiator interface {
	Negotiate(ctx context.Context, conn Conn) (context.Context, Conn, error)
	Name() string
}
