package fabric

import (
	"context"
)

type HandlerFunc func(ctx context.Context, conn Conn) (err error)

type Handler interface {
	Handle(ctx context.Context, conn Conn) (err error)
	CanHandle(addr string) bool
}
