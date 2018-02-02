package fabric

import (
	"context"
)

// Middleware are composites of a handler, a negotiator, and a name methods
type Middleware interface {
	Handle(ctx context.Context, conn Conn) (context.Context, Conn, error)
	Negotiate(ctx context.Context, conn Conn) (context.Context, Conn, error)
	Name() string
}
