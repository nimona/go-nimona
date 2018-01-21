package fabric

import (
	"context"
	"crypto/tls"
)

// SecMiddleware is a TLS middleware
type SecMiddleware struct {
	Config tls.Config
}

// Name of the middleware
func (m *SecMiddleware) Name() string {
	return "tls"

// Handle is the middleware handler for the server
func (m *SecMiddleware) Handle(ctx context.Context, c Conn) (context.Context, Conn, error) {
	scon := tls.Server(c, &m.Config)
	fmt.Println("tls, hnd")
	if err := scon.Handshake(); err != nil {
		return nil, nil, err
	}
	fmt.Println("tls, hnd-ed")

	nc := newConnWrapper(scon, c.GetAddress())
	fmt.Println("tls, neg-ed")
	return ctx, nc, nil
}

// Negotiate handles the client's side of the tls middleware
func (m *SecMiddleware) Negotiate(ctx context.Context, c Conn) (context.Context, Conn, error) {
	scon := tls.Client(c, &m.Config)
	fmt.Println("tls, neg")
	if err := scon.Handshake(); err != nil {
		return ctx, nil, err
	}
	fmt.Println("tls, neg-ed")

	nc := newConnWrapper(scon, c.GetAddress())
	fmt.Println("tls, ncw")
	return ctx, nc, nil
}
