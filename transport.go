package fabric

import (
	"context"
)

// Transport for dialing
type Transport interface {
	DialContext(ctx context.Context, addr *Address) (context.Context, Conn, error)
	CanDial(addr *Address) (bool, error)
	Listen(context.Context, HandlerFunc) error
	Addresses() []string
}
