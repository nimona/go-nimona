package dht

import (
	"context"

	peer "github.com/nimona/go-nimona-fabric/peer"
)

type DHT interface {
	Get(ctx context.Context, key string) (chan string, error)
	GetPeer(ctx context.Context, id string) (peer.Peer, error)
	GetLocalPairs() (map[string][]*Pair, error)
}
