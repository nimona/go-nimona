package fabric

import (
	"context"
	"crypto/tls"
	"fmt"
)

const (
	SecKey = "tls"
)

type SecMiddleware struct {
	Config tls.Config
}

func (m *SecMiddleware) Wrap(f HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, ucon Conn) error {
		fmt.Println("Going through sec")
		rc, err := ucon.GetRawConn()
		if err != nil {
			return err
		}

		scon := tls.Server(rc, &m.Config)
		if err := scon.Handshake(); err != nil {
			return err
		}

		if err := ucon.Upgrade(scon); err != nil {
			return err
		}

		return f(ctx, ucon)
	}
}

func (m *SecMiddleware) CanHandle(addr string) bool {
	parts := addrSplit(addr)
	return parts[0][0] == SecKey
}

func (m *SecMiddleware) Negotiate(ctx context.Context, ucon Conn) error {
	rc, err := ucon.GetRawConn()
	if err != nil {
		return err
	}

	scon := tls.Client(rc, &m.Config)
	if err := scon.Handshake(); err != nil {
		return err
	}

	return ucon.Upgrade(scon)
}

func (m *SecMiddleware) CanNegotiate(addr string) bool {
	parts := addrSplit(addr)
	return parts[0][0] == SecKey
}
