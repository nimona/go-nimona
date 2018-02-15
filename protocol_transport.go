package fabric

import "context"

type transportWrapper struct {
	protocolNames []string
}

// Handle adds the base protocols for transports
func (m *transportWrapper) Handle(fn HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c Conn) error {
		wrpAddr := c.GetAddress()
		wrpAddr.stack = append(wrpAddr.stack, m.protocolNames...)
		wrpConn := newConnWrapper(c, wrpAddr)
		return fn(ctx, wrpConn)
	}
}

// Negotiate is empty
func (m *transportWrapper) Negotiate(fn NegotiatorFunc) NegotiatorFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c Conn) error {
		return fn(ctx, c)
	}
}

// Name of the protocol
func (m *transportWrapper) Name() string {
	return ""
}
