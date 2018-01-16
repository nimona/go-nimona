package fabric

import (
	"context"

	"github.com/hashicorp/yamux"
)

// YamuxMiddleware is a multiplexer middleware based on yamux
type YamuxMiddleware struct{}

// HandlerWrapper is the middleware handler for the server
func (m *YamuxMiddleware) HandlerWrapper(f HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c Conn) error {
		session, err := yamux.Server(c, nil)
		if err != nil {
			return err
		}

		stream, err := session.Accept()
		if err != nil {
			return err
		}

		nc := newConnWrapper(stream)

		return f(ctx, nc)
	}
}

// Negotiate handles the client's side of the yamux middleware
func (m *YamuxMiddleware) Negotiate(ctx context.Context, c Conn) (context.Context, Conn, error) {
	session, err := yamux.Client(c, nil)
	if err != nil {
		return ctx, nil, err
	}

	stream, err := session.Open()
	if err != nil {
		return ctx, nil, err
	}

	nc := newConnWrapper(stream)

	return ctx, nc, nil
}
