package fabric

import (
	"context"
	"net"
)

type Transport interface {
	DialContext(ctx context.Context, addr string) (net.Conn, error)
	CanDial(addr string) bool
}
