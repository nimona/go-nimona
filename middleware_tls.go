package fabric

import (
	"context"
	"crypto/tls"
)

// SecMiddleware is a TLS middlware
type SecMiddleware struct {
	Config tls.Config
}

// HandlerWrapper is the middleware handler for the server
func (m *SecMiddleware) HandlerWrapper(f HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c Conn) error {
		scon := tls.Server(c, &m.Config)
		if err := scon.Handshake(); err != nil {
			return err
		}

		nc := newConnWrapper(scon)

		return f(ctx, nc)
	}
}

// Negotiate handles the client's side of the tls middleware
func (m *SecMiddleware) Negotiate(ctx context.Context, c Conn) (context.Context, Conn, error) {
	scon := tls.Client(c, &m.Config)
	if err := scon.Handshake(); err != nil {
		return ctx, nil, err
	}

	nc := newConnWrapper(scon)

	return ctx, nc, nil
}
