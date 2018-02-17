package transport

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	upnp "github.com/NebulousLabs/go-upnp"
	zap "go.uber.org/zap"

	address "github.com/nimona/go-nimona-fabric/address"
	conn "github.com/nimona/go-nimona-fabric/connection"
	logging "github.com/nimona/go-nimona-fabric/logging"
	protocol "github.com/nimona/go-nimona-fabric/protocol"
)

// NewTransportTCP returns a new TCP transport
func NewTransportTCP(host string, port int) Transport {
	return &TCP{
		host: host,
		port: port,
		upnp: nil,
	}
}

// NewTransportTCPWithUPNP returns a new TCP transport
func NewTransportTCPWithUPNP(host string, port int) Transport {
	upnp, _ := upnp.Discover()
	// TODO log error

	return &TCP{
		host: host,
		port: port,
		upnp: upnp,
	}
}

// TCP transport
type TCP struct {
	host     string
	port     int
	listener net.Listener
	upnp     UPNP
}

// DialContext attemps to dial to the peer with the given addr
func (t *TCP) DialContext(ctx context.Context, addr *address.Address) (context.Context, conn.Conn, error) {
	pr := addr.CurrentParams()
	tcon, err := net.Dial("tcp", pr)
	if err != nil {
		return nil, nil, err
	}

	addr.Pop()

	c := conn.NewConnWrapper(tcon, addr)

	return ctx, c, nil
}

// CanDial checks if address can be dialed by this transport
func (t *TCP) CanDial(addr *address.Address) (bool, error) {
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
func (t *TCP) Listen(ctx context.Context, handler protocol.HandlerFunc) error {
	addr := fmt.Sprintf("%s:%d", t.host, t.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	t.listener = listener

	err = t.startExternal(ctx)
	if err != nil {
		// Logger(ctx).Debug("Could not open upnp port: ", zap.Error(err))
	}

	go func() {
		for {
			// Listen for an incoming connection.
			tcon, err := listener.Accept()
			addr := address.NewAddress(addr)
			addr.Pop()
			c := conn.NewConnWrapper(tcon, addr)
			if err != nil {
				logging.Logger(ctx).Error("Could not accept TCP connection",
					zap.Error(err))
				continue
			}
			go t.handleListen(ctx, c, handler)
		}
	}()

	return nil
}

func (t *TCP) handleListen(ctx context.Context, conn conn.Conn, handler protocol.HandlerFunc) {
	if err := handler(ctx, conn); err != nil {
		logging.Logger(ctx).Error("Listen: Could not handle request",
			zap.Error(err))
	}
}

func (t *TCP) startExternal(ctx context.Context) error {
	if t.upnp == nil {
		return nil
	}

	_, xpstr, err := net.SplitHostPort(t.listener.Addr().String())
	if err != nil {
		return err
	}

	extPort, err := strconv.ParseUint(xpstr, 10, 16)
	if err != nil {
		return err
	}

	err = t.upnp.Clear(uint16(extPort))
	if err != nil {
		logging.Logger(ctx).Debug("Could not clear upnp: ", zap.Error(err))
	}

	err = t.upnp.Forward(uint16(extPort), "fabric")
	if err != nil {
		return err
	}

	return nil
}

// Addresses returns the addresses the transport is listening to
func (t *TCP) Addresses() []string {
	port := t.listener.Addr().(*net.TCPAddr).Port
	// TODO log errors
	addrs, _ := GetLocalAddresses(port)
	publicAddrs, _ := GetPublicAddresses(port, t.upnp)
	addrs = append(addrs, publicAddrs...)
	for i, addr := range addrs {
		addrs[i] = "tcp:" + addr
	}
	return addrs
}
