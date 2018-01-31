package fabric

import (
	"context"
	"net"
)

// Transport for dialing
type Transport interface {
	DialContext(ctx context.Context, addr Address) (net.Conn, error)
	CanDial(addr Address) (bool, error)
	Listen(context.Context, func(context.Context, net.Conn) error) error
	Addresses() []string
}
