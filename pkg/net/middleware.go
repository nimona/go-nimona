package net

import (
	"nimona.io/pkg/context"
)

// MiddlewareHandler ...
type MiddlewareHandler func(ctx context.Context,
	conn *Connection) (*Connection, error)

// Middleware ...
type Middleware interface {
	Handle() MiddlewareHandler
}
