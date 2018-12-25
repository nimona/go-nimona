package net

import "nimona.io/go/peers"

// ResolverProvider defines the interface for a resolver provider, eg our DHT
type ResolverProvider interface {
	Resolve(q *peers.PeerInfoRequest) ([]*peers.PeerInfo, error)
}
