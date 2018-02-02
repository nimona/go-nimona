package fabric

import "context"

// HandlerFunc defines the handler function for the server
type HandlerFunc func(ctx context.Context, conn Conn) (context.Context, Conn, error)

// Handler is responsible for handling a negotiation on the server's side
type Handler interface {
	Handle(ctx context.Context, conn Conn) (context.Context, Conn, error)
	Name() string
}
