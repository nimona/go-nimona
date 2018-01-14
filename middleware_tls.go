package fabric

import (
	"context"
	"crypto/tls"
)

const (
	SecKey = "tls"
)

type SecMiddleware struct {
	Config tls.Config
}

func (m *SecMiddleware) Wrap(f HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c Conn) error {
		scon := tls.Server(c, &m.Config)
		if err := scon.Handshake(); err != nil {
			return err
		}

		nc := newConnWrapper(scon, c.(*conn).remainingStack())

		return f(ctx, nc)
	}
}

func (m *SecMiddleware) Negotiate(ctx context.Context, c Conn) (Conn, error) {
	scon := tls.Client(c, &m.Config)
	if err := scon.Handshake(); err != nil {
		return nil, err
	}

	nc := newConnWrapper(scon, c.(*conn).remainingStack())

	return nc, nil
}
