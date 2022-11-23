package nimona

import (
	"context"
	"fmt"
	"net"

	"github.com/neilalexander/utp"
)

type TransportUTP struct{}

func (t *TransportUTP) Dial(
	ctx context.Context,
	addr NodeAddr,
) (net.Conn, error) {
	if addr.Network() != "utp" {
		return nil, ErrTransportUnsupported
	}

	c, err := utp.DialContext(ctx, addr.Address())
	if err != nil {
		return nil, fmt.Errorf("utp: failed to dial: %w", err)
	}
	return c, nil
}

func (t *TransportUTP) Listen(addr string) (Listener, error) {
	l, err := utp.Listen(addr)
	if err != nil {
		return nil, fmt.Errorf("utp: failed to listen: %w", err)
	}

	return wrapListener(l, "utp"), nil
}
