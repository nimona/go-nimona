package fabric

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"golang.org/x/net/websocket"
)

// NewTransportWebsocket returns a new Websocket transport
func NewTransportWebsocket(host string, port int) Transport {
	return &Websocket{
		host: host,
		port: port,
	}
}

// Websocket transport
type Websocket struct {
	host     string
	port     int
	listener net.Listener
}

// Name of the transport
func (t *Websocket) Name() string {
	return "ws"
}

// CanDial checks if address can be dialed by this transport
func (t *Websocket) CanDial(addr Address) (bool, error) {
	if addr.CurrentProtocol() != "ws" {
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

// DialContext attempts to dial to the peer with the given address
func (t *Websocket) DialContext(ctx context.Context, addr Address) (
	net.Conn, error) {
	pr := addr.CurrentParams()

	// TODO fix origin to use a real address
	origin := fmt.Sprintf("ws://%s:%d", t.host, t.port)
	tcon, err := websocket.Dial("ws://"+pr, "", origin)
	if err != nil {
		return nil, err
	}

	return tcon, nil
}

// Listen starts listening for incoming connections
func (t *Websocket) Listen(ctx context.Context, handler func(context.Context, net.Conn) error) error {
	lgr := Logger(ctx)

	wsh := websocket.Handler(func(conn *websocket.Conn) {
		if err := handler(ctx, conn); err != nil {
			lgr.Error("Could not handle ws connection", zap.Error(err))
		}
	})

	addr := fmt.Sprintf("%s:%d", t.host, t.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	t.listener = listener

	go func() {
		if err := http.Serve(listener, wsh); err != nil {
			lgr.Error("Could not listen for ws connection", zap.Error(err))
		}
	}()

	return nil
}

// Addresses returns the address the transport is listening to
func (t *Websocket) Addresses() []string {
	port := t.listener.Addr().(*net.TCPAddr).Port
	addrs, err := GetAddresses(port)
	if err != nil {
		return []string{}
	}

	for i, addr := range addrs {
		addrs[i] = "ws:" + addr
	}

	return addrs
}
