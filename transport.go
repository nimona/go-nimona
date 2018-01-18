package fabric

import (
	"context"
	"net"
)

type Transport interface {
	DialContext(ctx context.Context, addr Address) (net.Conn, error)
	CanDial(addr Address) (bool, error)
}
