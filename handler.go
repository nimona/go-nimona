package fabric

import (
	"context"
)

type HandlerFunc func(ctx context.Context, conn Conn) (err error)

type Handler interface {
	Wrap(HandlerFunc) HandlerFunc
	CanHandle(addr string) bool
}
