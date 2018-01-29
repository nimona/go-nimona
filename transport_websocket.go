package fabric

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

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
	return "websocket"
}

// CanDial checks if address can be dialed by this transport
func (t *Websocket) CanDial(addr Address) (bool, error) {
	return addr.CurrentProtocol() == "websocket", nil
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
func (t *Websocket) Listen(handler func(net.Conn) error) error {
	http.Handle("/", websocket.Handler(func(conn *websocket.Conn) {
		defer func() {
			if err := conn.Close(); err != nil {
				fmt.Println("Could not close conn", err)
			}
		}()
		if err := handler(conn); err != nil {
			fmt.Println("Listen: Could not handle request. error:", err)
		}
	}))

	err := http.ListenAndServe(t.address, nil)
	if err != nil {
		return err
	}

	return nil
}

// Address returns the address the transport is listening to
func (t *Websocket) Address() string {
	return t.address
}
