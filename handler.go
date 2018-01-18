package fabric

import (
	"context"
)

// Handler defines the handler function for the server
type Handler func(context.Context, Conn) (context.Context, Conn, error)
