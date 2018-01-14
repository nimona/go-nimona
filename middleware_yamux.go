package fabric

import (
	"context"

	"github.com/hashicorp/yamux"
)

const (
	YamuxKey = "yamux"
)

type YamuxMiddleware struct{}

func (m *YamuxMiddleware) Wrap(f HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, ucon Conn) error {
		rc, err := ucon.GetRawConn()
		if err != nil {
			return err
		}

		session, err := yamux.Server(rc, nil)
		if err != nil {
			return err
		}

		stream, err := session.Accept()
		if err != nil {
			return err
		}

		if err := ucon.Upgrade(stream); err != nil {
			return err
		}

		return f(ctx, ucon)
	}
}

func (m *YamuxMiddleware) Negotiate(ctx context.Context, ucon Conn) error {
	rc, err := ucon.GetRawConn()
	if err != nil {
		return err
	}

	session, err := yamux.Client(rc, nil)
	if err != nil {
		return err
	}

	stream, err := session.Open()
	if err != nil {
		return err
	}

	return ucon.Upgrade(stream)
}
