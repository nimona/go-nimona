package fabric

import (
	"context"
)

type Handler interface {
	Handle(ctx context.Context, conn Conn) (newConn Conn, err error)
	CanHandle(addr string) bool
}

type Negotiator interface {
	Negotiate(ctx context.Context, conn Conn) (newConn Conn, err error)
	CanNegotiate(addr string) bool
}

type Middleware interface {
	Handler
	Negotiator
}
