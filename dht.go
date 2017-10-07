package dht

import (
	"context"

	net "github.com/nimona/go-nimona-net"
)

// DHT ..
type DHT interface {
	Get(ctx context.Context, key string) (chan string, error)
	GetPeer(ctx context.Context, id string) (net.Peer, error)
	GetLocalPairs() (map[string][]*Pair, error)
}
