package fabric

import (
	"context"

	"github.com/hashicorp/yamux"
)

const (
	YamuxKey = "yamux"
)

type YamuxMiddleware struct{}

func (m *YamuxMiddleware) Handle(ctx context.Context, ucon Conn) error {
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

	return ucon.Upgrade(stream)
}

func (m *YamuxMiddleware) CanHandle(addr string) bool {
	parts := addrSplit(addr)
	return parts[0][0] == YamuxKey
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

func (m *YamuxMiddleware) CanNegotiate(addr string) bool {
	parts := addrSplit(addr)
	return parts[0][0] == YamuxKey
}
