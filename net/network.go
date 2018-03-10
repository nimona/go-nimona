package fabric

import (
	"context"
)

// Network is our network interface for fabric
type Network interface {
	DialContext(ctx context.Context, as string) (context.Context, Conn, error)
	AddTransport(transport Transport, protocols []Protocol) error
	AddProtocol(protocol Protocol) error
	GetAddresses() []string
}
