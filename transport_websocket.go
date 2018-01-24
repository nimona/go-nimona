package fabric

import (
	"context"
	"net"
)

// NewTransportWebsocket returns a new Websocket transport
func NewTransportWebsocket(addr string) Transport {
	return &Websocket{address: addr}
}

// Websocket transport
type Websocket struct {
	address string
}

// Name of the transport
func (t *Websocket) Name() string {
	return "websocket"
}

// CanDial checks if address can be dialed by this transport
func (t *Websocket) CanDial(addr Address) (bool, error) {
	return addr.CurrentProtocol() == "websocket", nil
}

// DialContext attempts to dial to the peer with the given address
func (t *Websocket) DialContext(ctx context.Context, addr Address) (
	net.Conn, error) {
	return nil, nil
}

// Listen starts listening for incoming connections
func (t *Websocket) Listen(handler func(net.Conn) error) error {
	return nil
}

// Address returns the address the transport is listening to
func (t *Websocket) Address() string {
	return t.address
}
