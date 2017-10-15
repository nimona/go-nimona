package fabric

import "context"

type Middleware interface {
	Handle(ctx context.Context, conn Conn) (Conn, error)
	Negotiate(ctx context.Context, conn Conn, param string) (Conn, error)
}
