package fabric

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"go.uber.org/zap"
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
	if addr.CurrentProtocol() != "tcp" {
		return false, nil
	}

	params := strings.Split(addr.CurrentParams(), ":")
	if len(params) != 2 {
		return false, errors.New("Invalid number of parameters")
	}

	if len(params[0]) == 0 {
		return false, errors.New("Missing destination host/ip")
	}

	port, err := strconv.Atoi(params[1])
	if err != nil {
		return false, errors.New("Invalid port")
	}

	if port == 0 || port > 65535 {
		return false, errors.New("Invalid port")
	}

	return true, nil
}

// Listen handles the transports
func (t *TCP) Listen(ctx context.Context, handler func(context.Context, net.Conn) error) error {
	// TODO read the address from the struct
	l, err := net.Listen("tcp", t.address)
	if err != nil {
		return err
	}

	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				Logger(ctx).Error("Could not accept TCP connection", zap.Error(err))
				continue
			}
			go t.handleListen(ctx, conn, handler)
		}
	}()

	return nil
}

func (t *TCP) handleListen(ctx context.Context, conn net.Conn, handler func(context.Context, net.Conn) error) {
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Println("Could not close conn", err)
		}
	}()
	if err := handler(ctx, conn); err != nil {
		fmt.Println("Listen: Could not handle request. error:", err)
	}
}

// Address returns the address the transport is listening to
func (t *TCP) Address() string {
	return t.address
}
