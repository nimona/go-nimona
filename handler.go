package fabric

import (
	"context"
)

// HandlerFunc defines the handler function for the server
type HandlerFunc func(ctx context.Context, conn Conn) (err error)

// HandlerFuncWrapper is a middleware wrapper for the HandlerFunc
type HandlerFuncWrapper func(HandlerFunc) HandlerFunc
