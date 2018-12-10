package net

import (
	"context"

	"nimona.io/go/crypto"
	"nimona.io/go/peers"
)

// Discoverer interface for finding peers and providers
type Discoverer interface {
	// GetPeerInfo returns a PeerInfo given its ID
	GetPeerInfo(ctx context.Context, id string) (*peers.PeerInfo, error)
	// PutProviders adds a key of something we provide
	PutProviders(ctx context.Context, key string) error
	// GetProviders returns Peer IDs that provide a key
	GetProviders(ctx context.Context, key string) (chan *crypto.Key, error)
	// GetAllProviders returns all local known providers and the keys they provide
	GetAllProviders() (map[string][]string, error)
}
