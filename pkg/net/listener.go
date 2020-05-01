package net

import "net"

type (
	Listener interface {
		Close() error
		Addresses() []string
	}
	listener struct {
		listeners []net.Listener
		addresses []string
	}
)

func (l *listener) Close() error {
	// TODO add multierror
	for _, lst := range l.listeners {
		lst.Close() // nolint: noerr
	}
	return nil
}

func (l *listener) Addresses() []string {
	return l.addresses
}
