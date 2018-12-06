package net

import (
	"errors"
	"sync"

	"nimona.io/go/peers"
)

// Resolver interface
type Resolver interface {
	AddProvider(provider ResolverProvider) error
	Resolve(key string) (*peers.PeerInfo, error)
	Add(v *peers.PeerInfo)
	// AddPersistent(v *peers.PeerInfo)
}

// NewResolver creates a new empty resolver with no providers
func NewResolver() Resolver {
	return &resolver{
		providersLock:   sync.RWMutex{},
		providers:       []ResolverProvider{},
		cacheLock:       sync.RWMutex{},
		cacheTemp:       map[string]*peers.PeerInfo{},
		cachePersistent: map[string]*peers.PeerInfo{},
	}
}

// resolver wrapps multiple providers to allow resolving peer keys to peer infos
// TODO consider allowing the resolver to accept an interface, and select
// the provider based on the input's type. This would require registering
// providers with the inputs they accept.
type resolver struct {
	providersLock   sync.RWMutex
	providers       []ResolverProvider
	cacheLock       sync.RWMutex
	cacheTemp       map[string]*peers.PeerInfo
	cachePersistent map[string]*peers.PeerInfo
}

// Resolve goes through the given providers until one returns something
func (r *resolver) Resolve(key string) (*peers.PeerInfo, error) {
	r.providersLock.RLock()
	for _, p := range r.providers {
		if res, err := p.Resolve(key); err == nil {
			r.providersLock.RUnlock()
			return res, nil
		}
	}
	r.providersLock.RUnlock()
	r.cacheLock.RLock()
	defer r.cacheLock.RUnlock()
	if res, ok := r.cacheTemp[key]; ok {
		return res, nil
	}
	if res, ok := r.cachePersistent[key]; ok {
		return res, nil
	}
	return nil, errors.New("could not resolve")
}

// AddProvider to the resolver
func (r *resolver) AddProvider(provider ResolverProvider) error {
	r.providersLock.Lock()
	r.providers = append(r.providers, provider)
	r.providersLock.Unlock()
	return nil
}

// Add allows manually adding peer infos to be resolved.
// These peers will eventually be gc-ed.
func (r *resolver) Add(v *peers.PeerInfo) {
	r.cacheLock.Lock()
	r.cacheTemp[v.HashBase58()] = v
	r.cacheLock.Unlock()
}

// AddPersistent allows adding permanent peer infos to be resolved.
// These peers can be overshadowed by other resolvers, but will never be gc-ed.
// Mainly used for adding bootstrap nodes.
func (r *resolver) AddPersistent(v *peers.PeerInfo) {
	r.cacheLock.Lock()
	r.cachePersistent[v.HashBase58()] = v
	r.cacheLock.Unlock()
}
