package fabric

import (
	"context"
)

// Protocol deals with both sides, server and client
type Protocol interface {
	Handle(ctx context.Context, conn Conn) (context.Context, Conn, error)
	Negotiate(ctx context.Context, conn Conn) (context.Context, Conn, error)
	Name() string
}
