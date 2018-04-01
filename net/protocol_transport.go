package net

import (
	"context"
)

// TransportWrapper for wrapping the first transport
// TODO better docs
type TransportWrapper struct {
	protocolNames []string
}

// NewTransportWrapper returns a new TransportWrapper
func NewTransportWrapper(protocolNames []string) *TransportWrapper {
	return &TransportWrapper{
		protocolNames: protocolNames,
	}
}

// Handle adds the base protocols for transports
func (m *TransportWrapper) Handle(fn HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c Conn) error {
		wrpAddr := c.GetAddress()
		wrpAddr.Append(m.protocolNames...)
		wrpConn := NewConnWrapper(c, wrpAddr)
		return fn(ctx, wrpConn)
	}
}

// Negotiate is empty
func (m *TransportWrapper) Negotiate(fn NegotiatorFunc) NegotiatorFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c Conn) error {
		return fn(ctx, c)
	}
}

// Name of the protocol
func (m *TransportWrapper) Name() string {
	return ""
}

// GetAddresses
func (m *TransportWrapper) GetAddresses() []string {
	return []string{}
}
