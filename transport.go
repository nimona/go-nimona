package nimona

import (
	"context"
	"net"
)

type (
	// Transport is a combination of a Dialer and a Listener
	Transport interface {
		Dialer
		Listen() (net.Listener, error)
	}
	// Dialer is an interface for dialing a connection
	Dialer interface {
		Dial(context.Context, NodeAddr) (net.Conn, error)
	}
	// Listener is an interface for listening for connections
	Listener interface {
		Accept() (net.Conn, error)
		Close() error
		NodeAddr() NodeAddr
	}
)

type listener struct {
	net.Listener
	transport string
}

func (l *listener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (l *listener) Close() error {
	return l.Listener.Close()
}

func (l *listener) NodeAddr() NodeAddr {
	return NodeAddr{
		Network: l.transport,
		Address: l.Listener.Addr().String(),
	}
}

func wrapListener(l net.Listener, transport string) Listener {
	return &listener{
		Listener:  l,
		transport: transport,
	}
}
