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

func (m *SecMiddleware) Handle(ctx context.Context, ucon Conn) (Conn, error) {
	scon := tls.Server(ucon, &m.Config)
	if err := scon.Handshake(); err != nil {
		return nil, err
	}

	conn := wrapConn(scon)
	return conn, nil
}

func (m *SecMiddleware) CanHandle(addr string) bool {
	parts := addrSplit(addr)
	return parts[0][0] == SecKey
}

func (m *SecMiddleware) Negotiate(ctx context.Context, ucon Conn) (Conn, error) {
	scon := tls.Client(ucon, &m.Config)
	if err := scon.Handshake(); err != nil {
		return nil, err
	}

	conn := wrapConn(scon)
	return conn, nil
}

func (m *SecMiddleware) CanNegotiate(addr string) bool {
	parts := addrSplit(addr)
	return parts[0][0] == SecKey
}
