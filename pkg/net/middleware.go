package net

import (
	"context"
)

// MiddlewareHandler ...
type MiddlewareHandler func(ctx context.Context,
	conn *Connection) (*Connection, error)

// Middleware ...
type Middleware interface {
	Handle() MiddlewareHandler
}
