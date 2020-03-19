package net

import (
	"nimona.io/pkg/context"
)

// MiddlewareHandler defines a middleware handler
type MiddlewareHandler func(ctx context.Context,
	conn *Connection) (*Connection, error)

// Middleware defines a middleware interface
type Middleware interface {
	Handle() MiddlewareHandler
}
