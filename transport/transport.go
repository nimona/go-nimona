package transport

import (
	"context"

	address "github.com/nimona/go-nimona-fabric/address"
	conn "github.com/nimona/go-nimona-fabric/connection"
	protocol "github.com/nimona/go-nimona-fabric/protocol"
)

// Transport for dialing
type Transport interface {
	DialContext(ctx context.Context, addr *address.Address) (context.Context, conn.Conn, error)
	CanDial(addr *address.Address) (bool, error)
	Listen(context.Context, protocol.HandlerFunc) error
	Addresses() []string
}
