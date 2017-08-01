package dht

import (
	"context"

	net "github.com/nimona/go-nimona-net"
)

// DHT ..
type DHT interface {
	Find(context.Context, string) (net.Peer, error)
	Ping(context.Context, net.Peer) (net.Peer, error)
}
