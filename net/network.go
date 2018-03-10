package net

import (
	"context"
)

// Net is our network interface for net
type Net interface {
	DialContext(ctx context.Context, as string) (context.Context, Conn, error)
	AddTransport(transport Transport, protocols []Protocol) error
	AddProtocol(protocol Protocol) error
	GetAddresses() []string
	CallContext(ctx context.Context, as string, extraProtocols ...Protocol) error
	DialAndProcessWithContext(ctx context.Context, as string, extraProtocols ...Protocol) (context.Context, Conn, error)
}
