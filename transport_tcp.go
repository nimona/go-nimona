package fabric

import (
	"context"
	"net"
)

// NewTransportTCP returns a new TCP transport
func NewTransportTCP() Transport {
	return &TCP{}
}

// TCP transport
type TCP struct{}

// DialContext attemps to dial to the peer with the given addr
func (t *TCP) DialContext(ctx context.Context, addr Address) (net.Conn, error) {
	pr := addr.CurrentParams()
	tcon, err := net.Dial("tcp", pr)
	if err != nil {
		return nil, err
	}

	return tcon, nil
}

// CanDial checks if address can be dialed by this transport
func (t *TCP) CanDial(addr Address) (bool, error) {
	return addr.CurrentProtocol() == "tcp", nil
}
