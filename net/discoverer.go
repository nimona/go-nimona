package net

import (
	"context"

	"github.com/nimona/go-nimona/peers"
)

type Discoverer interface {
	GetPeerInfo(ctx context.Context, id string) (*peers.PeerInfo, error)
	PutProviders(ctx context.Context, key string) error
	GetProviders(ctx context.Context, key string) ([]string, error)
	// TODO do we need those?
	GetAllProviders() (map[string][]string, error)
}
