package net

import (
	"nimona.io/internal/context"
)

// MiddlewareHandler ...
type MiddlewareHandler func(ctx context.Context,
	conn *Connection) (*Connection, error)

// Middleware ...
type Middleware interface {
	Handle() MiddlewareHandler
}
