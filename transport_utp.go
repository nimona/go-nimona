package nimona

import (
	"context"
	"fmt"
	"net"

	"github.com/neilalexander/utp"
)

var ErrTransportUnsupported error = fmt.Errorf("transport unsupported")

// TransportUTP is a transport that uses uTP, a UDP based protocol
// for reliable data transfer.
type TransportUTP struct {
	PublicAddress string
	PublicKey     PublicKey
}

func (t *TransportUTP) Dial(ctx context.Context, addr PeerAddr) (net.Conn, error) {
	if addr.Network != "utp" {
		return nil, ErrTransportUnsupported
	}

	c, err := utp.DialContext(ctx, addr.Address)
	if err != nil {
		return nil, fmt.Errorf("utp: failed to dial: %w", err)
	}
	return c, nil
}

func (t *TransportUTP) Listen(ctx context.Context, addr string) (Listener, error) {
	l, err := utp.NewSocket("udp4", addr)
	if err != nil {
		return nil, fmt.Errorf("utp: failed to listen: %w", err)
	}

	return wrapListener(l, "utp", t.PublicAddress, t.PublicKey), nil
}
