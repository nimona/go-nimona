package dht

import "context"

// DHT ..
type DHT interface {
	Find(context.Context, ID) (Peer, error)
	Ping(context.Context, Peer) (Peer, error)
}
