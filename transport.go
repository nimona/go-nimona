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
		Dial(context.Context, PeerAddr) (net.Conn, error)
	}
	// Listener is an interface for listening for connections
	Listener interface {
		Accept() (net.Conn, error)
		Close() error
		PeerAddr() PeerAddr
	}
)

type listener struct {
	net.Listener
	addr PeerAddr
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

func (l *listener) PeerAddr() PeerAddr {
	return l.addr
}

func wrapListener(l net.Listener, transport, publicHost string, publicKey PublicKey) Listener {
	host := l.Addr().String()
	if publicHost != "" {
		_, port, _ := net.SplitHostPort(host)
		host = net.JoinHostPort(publicHost, port)
	}
	return &listener{
		Listener: l,
		addr: PeerAddr{
			Network:   transport,
			Address:   host,
			PublicKey: publicKey,
		},
	}
}
