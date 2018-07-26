package net

import "context"

type Discoverer interface {
	GetPeerInfo(ctx context.Context, id string) (*PeerInfo, error)
	PutValue(ctx context.Context, key, value string) error
	GetValue(ctx context.Context, key string) (string, error)
	PutProviders(ctx context.Context, key string) error
	GetProviders(ctx context.Context, key string) ([]string, error)
	// TODO do we need those?
	GetAllProviders() (map[string][]string, error)
	GetAllValues() (map[string]string, error)
}
