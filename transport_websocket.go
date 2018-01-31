package fabric

import (
	"context"
	"net"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"golang.org/x/net/websocket"
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
	return "ws"
}

// CanDial checks if address can be dialed by this transport
func (t *Websocket) CanDial(addr Address) (bool, error) {
	return addr.CurrentProtocol() == "ws", nil
}

// DialContext attempts to dial to the peer with the given address
func (t *Websocket) DialContext(ctx context.Context, addr Address) (
	net.Conn, error) {
	pr := addr.CurrentParams()

	tcon, err := websocket.Dial("ws://"+pr, "", "ws://"+strings.Split(t.address, ":")[0])
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

	go func() {
		if err := http.ListenAndServe(t.address, wsh); err != nil {
			lgr.Error("Could not listen for ws connection", zap.Error(err))
		}
	}()

	return nil
}

// Addresses returns the address the transport is listening to
func (t *Websocket) Addresses() []string {
	return []string{"ws:" + t.address}
}
