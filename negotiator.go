package fabric

import (
	"context"
)

// NegotiatorFunc defines the negotiator functions for the clients
type NegotiatorFunc func(ctx context.Context, conn Conn) (context.Context, Conn, error)
