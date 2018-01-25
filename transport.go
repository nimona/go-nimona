package fabric

import (
	"context"
	"net"
)

// Transport for dialing
type Transport interface {
	DialContext(ctx context.Context, addr Address) (net.Conn, error)
	CanDial(addr Address) (bool, error)
	Name() string
	Listen(context.Context, func(context.Context, net.Conn) error) error
	Address() string
}
