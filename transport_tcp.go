package fabric

import (
	"context"
	"fmt"
	"net"
)

// NewTransportTCP returns a new TCP transport
func NewTransportTCP(addr string) Transport {
	return &TCP{address: addr}
}

// TCP transport
type TCP struct {
	address string
}

// Name of the transport
func (t *TCP) Name() string {
	return "tcp"
}

// DialContext attemps to dial to the peer with the given addr
func (t *TCP) DialContext(ctx context.Context, addr Address) (
	net.Conn, error) {
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

// Listen handles the transports
func (t *TCP) Listen(handler func(net.Conn) error) error {
	// TODO read the address from the struct
	l, err := net.Listen("tcp", t.address)
	if err != nil {
		return err
	}

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go t.handleListen(conn, handler)
	}
}

func (t *TCP) handleListen(conn net.Conn, handler func(net.Conn) error) {

	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Println("Could not close conn", err)
		}
	}()
	if err := handler(conn); err != nil {
		fmt.Println("Listen: Could not handle request. error:", err)
	}
}

// Address returns the address the transport is listening to
func (t *TCP) Address() string {
	return t.address
}
