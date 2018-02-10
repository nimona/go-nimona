package fabric

import (
	"context"
)

// Protocol are composites of a handler, a negotiator, and a name methods
type Protocol interface {
	Handle(ctx context.Context, conn Conn) (context.Context, Conn, error)
	Negotiate(ctx context.Context, conn Conn) (context.Context, Conn, error)
	Name() string
}
