package discovery

import (
	"errors"
	"sync"

	"nimona.io/pkg/net/peer"
)

// nolint: lll
//go:generate go run github.com/vektra/mockery/cmd/mockery -name Discoverer -case underscore

// discovererOptions is the complete options structure for the discoverer
type discovererOptions struct {
	Local bool
}

// DiscovererOption is the type for our functional options
type DiscovererOption func(*discovererOptions)

// Local forces the discoverer to only look at its cache
func Local() DiscovererOption {
	return func(opts *discovererOptions) {
		opts.Local = true
	}
}

func parseSendOptions(opts ...DiscovererOption) *discovererOptions {
	options := &discovererOptions{}
	for _, o := range opts {
		o(options)
	}
	return options
}

// Discoverer interface
type Discoverer interface {
	AddProvider(provider Provider) error
	Discover(q *peer.PeerInfoRequest, options ...DiscovererOption) ([]*peer.PeerInfo, error)
	Add(v *peer.PeerInfo)
	// AddPersistent(v *peer.PeerInfo)
}

// NewDiscoverer creates a new empty discoverer with no providers
func NewDiscoverer() Discoverer {
	return &discoverer{
		providersLock:   sync.RWMutex{},
		providers:       []Provider{},
		cacheLock:       sync.RWMutex{},
		cacheTemp:       map[string]*peer.PeerInfo{},
		cachePersistent: map[string]*peer.PeerInfo{},
	}
}

// discoverer wraps multiple providers to allow resolving peer keys to peer infos
// TODO consider allowing the discoverer to accept an interface, and select
// the provider based on the input's type. This would require registering
// providers with the inputs they accept.
type discoverer struct {
	providersLock   sync.RWMutex
	providers       []Provider
	cacheLock       sync.RWMutex
	cacheTemp       map[string]*peer.PeerInfo
	cachePersistent map[string]*peer.PeerInfo
}

// Discover goes through the given providers until one returns something
func (r *discoverer) Discover(q *peer.PeerInfoRequest, opts ...DiscovererOption) ([]*peer.PeerInfo, error) {
	cfg := parseSendOptions(opts...)
	if !cfg.Local {
		r.providersLock.RLock()
		for _, p := range r.providers {
			if res, err := p.Discover(q); err == nil && res != nil {
				r.providersLock.RUnlock()
				return res, nil
			}
		}
		r.providersLock.RUnlock()
	}

	// we only cache peer infos by their peer id
	if q.SignerKeyHash == "" {
		return nil, errors.New("could not resolve")
	}

	r.cacheLock.RLock()
	defer r.cacheLock.RUnlock()
	if res, ok := r.cacheTemp[q.SignerKeyHash]; ok && res != nil {
		return []*peer.PeerInfo{res}, nil
	}

	if res, ok := r.cachePersistent[q.SignerKeyHash]; ok && res != nil {
		return []*peer.PeerInfo{res}, nil
	}

	return nil, errors.New("could not resolve")
}

// AddProvider to the discoverer
func (r *discoverer) AddProvider(provider Provider) error {
	r.providersLock.Lock()
	r.providers = append(r.providers, provider)
	r.providersLock.Unlock()
	return nil
}

// Add allows manually adding peer infos to be resolved.
// These peers will eventually be gc-ed.
func (r *discoverer) Add(v *peer.PeerInfo) {
	r.cacheLock.Lock()
	r.cacheTemp[v.HashBase58()] = v
	r.cacheLock.Unlock()
}

// AddPersistent allows adding permanent peer infos to be resolved.
// These peers can be overshadowed by other discoverers, but will never be gc-ed.
// Mainly used for adding bootstrap nodes.
func (r *discoverer) AddPersistent(v *peer.PeerInfo) {
	r.cacheLock.Lock()
	r.cachePersistent[v.HashBase58()] = v
	r.cacheLock.Unlock()
}
