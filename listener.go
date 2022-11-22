package nimona

import (
	"net"
	"strconv"
)

type (
	Listener interface {
		net.Listener
		NodeAddr() NodeAddr
	}
	listener struct {
		net.Listener
		transport string
	}
)

func (l *listener) NodeAddr() NodeAddr {
	host, port, _ := net.SplitHostPort(l.Listener.Addr().String())
	portInt, _ := strconv.Atoi(port)
	return NewNodeAddr(l.transport, host, portInt)
}

func wrapListener(l net.Listener, transport string) Listener {
	return &listener{
		Listener:  l,
		transport: transport,
	}
}
