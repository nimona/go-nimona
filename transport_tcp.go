package fabric

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"

	upnp "github.com/NebulousLabs/go-upnp"
	"go.uber.org/zap"
)

// NewTransportTCP returns a new TCP transport
func NewTransportTCP(host string, port int) Transport {
	return &TCP{
		host: host,
		port: port,
	}
}

// TCP transport
type TCP struct {
	host     string
	port     int
	listener net.Listener
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

	hostPort := addr.CurrentParams()
	_, portString, err := net.SplitHostPort(hostPort)
	if err != nil {
		return false, err
	}

	port, err := strconv.Atoi(portString)
	if err != nil || port == 0 || port > 65535 {
		return false, errors.New("Invalid port number")
	}

	return true, nil
}

// Listen handles the transports
func (t *TCP) Listen(ctx context.Context, handler func(context.Context, net.Conn) error) error {
	addr := fmt.Sprintf("%s:%d", t.host, t.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	t.listener = listener

	upr, err := upnp.Discover()
	if err != nil {
		log.Fatal("Router not found: ", err)
	}

	err = upr.Forward(uint16(t.port), "fabric")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := listener.Accept()
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
	if err := handler(ctx, conn); err != nil {
		fmt.Println("Listen: Could not handle request. error:", err)
	}
}

// Addresses returns the addresses the transport is listening to
func (t *TCP) Addresses() []string {
	port := t.listener.Addr().(*net.TCPAddr).Port
	addrs, err := GetAddresses(port)
	if err != nil {
		return []string{}
	}

	for i, addr := range addrs {
		addrs[i] = "tcp:" + addr
	}

	return addrs
}
