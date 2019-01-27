package discovery

import "nimona.io/pkg/net/peer"

// Provider defines the interface for a discoverer provider, eg our DHT
type Provider interface {
	Discover(q *peer.PeerInfoRequest) ([]*peer.PeerInfo, error)
}
