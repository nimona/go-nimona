package discovery

import (
	"fmt"
	"sync"

	"nimona.io/internal/context"
	"nimona.io/internal/errors"
	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/peer"
)

//go:generate $GOBIN/mockery -name Discoverer -case underscor
//go:generate $GOBIN/mockery -name Provider -case underscore

type (
	// Provider defines the interface for a discoverer provider, eg our DHT
	Provider interface {
		FindByFingerprint(
			ctx context.Context,
			fingerprint crypto.Fingerprint,
			opts ...Option,
		) ([]*peer.Peer, error)
		FindByContent(
			ctx context.Context,
			contentHash string,
			opts ...Option,
		) ([]crypto.Fingerprint, error)
	}
	// Discoverer interface
	Discoverer interface {
		AddProvider(provider Provider) error
		Add(peer *peer.Peer)
		FindByFingerprint(
			ctx context.Context,
			fingerprint crypto.Fingerprint,
			opts ...Option,
		) ([]*peer.Peer, error)
		FindByContent(
			ctx context.Context,
			contentHash string,
			opts ...Option,
		) ([]crypto.Fingerprint, error)
	}
)

// Options is the complete options structure for the discoverer
type Options struct {
	Local bool
}

// Option is the type for our functional options
type Option func(*Options)

// Local forces the discoverer to only look at its cache
func Local() Option {
	return func(opts *Options) {
		opts.Local = true
	}
}

func ParseOptions(opts ...Option) *Options {
	options := &Options{}
	for _, o := range opts {
		o(options)
	}
	return options
}

// NewDiscoverer creates a new empty discoverer with no providers
func NewDiscoverer() Discoverer {
	return &discoverer{
		providersLock:   sync.RWMutex{},
		providers:       []Provider{},
		cacheLock:       sync.RWMutex{},
		cacheTemp:       map[string]*peer.Peer{},
		cachePersistent: map[string]*peer.Peer{},
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
	cacheTemp       map[string]*peer.Peer
	cachePersistent map[string]*peer.Peer
}

// FindByFingerprint goes through the given providers until one returns something
func (r *discoverer) FindByFingerprint(
	ctx context.Context,
	fingerprint crypto.Fingerprint,
	opts ...Option,
) ([]*peer.Peer, error) {
	opt := ParseOptions(opts...)

	logger := log.FromContext(ctx).With(
		log.String("method", "discovery/discoverer.FindByFingerprint"),
		log.String("fingerprint", fingerprint.String()),
		log.String("opts", fmt.Sprintf("%#v", opt)),
	)

	logger.Debug("trying to find peers")

	// r.providersLock.RLock()
	// defer r.providersLock.RUnlock()

	ps := []*peer.Peer{}
	for _, p := range r.providers {
		eps, err := p.FindByFingerprint(ctx, fingerprint, opts...)
		if err != nil {
			logger.With(
				log.Error(err),
			).Debug("provider failed")
			continue
		}
		ps = append(ps, eps...)
		logger.With(
			log.Int("n", len(eps)),
			log.Any("peers", ps),
		).Debug("found n peers")
	}

	// r.cacheLock.RLock()
	// defer r.cacheLock.RUnlock()

	// TODO move persistence into its own provider

	if res, ok := r.cacheTemp[fingerprint.String()]; ok && res != nil {
		ps = append(ps, res)
	}

	if res, ok := r.cachePersistent[fingerprint.String()]; ok && res != nil {
		ps = append(ps, res)
	}

	if len(ps) == 0 {
		return nil, errors.New("could not resolve")
	}

	return ps, nil
}

// FindByContent goes through the given providers until one returns something
func (r *discoverer) FindByContent(
	ctx context.Context,
	contentHash string,
	opts ...Option,
) ([]crypto.Fingerprint, error) {
	opt := ParseOptions(opts...)

	logger := log.FromContext(ctx).With(
		log.String("method", "discovery/discoverer.FindByContent"),
		log.String("contentHash", contentHash),
		log.String("opts", fmt.Sprintf("%#v", opt)),
	)

	logger.Debug("trying to find peers")

	ps := []crypto.Fingerprint{}
	for _, p := range r.providers {
		eps, err := p.FindByContent(ctx, contentHash, opts...)
		if err != nil {
			logger.With(
				log.Error(err),
			).Debug("provider failed")
			continue
		}
		ps = append(ps, eps...)
		logger.With(
			log.Int("n", len(eps)),
			log.Any("peers", ps),
		).Debug("found n peers")
	}

	if len(ps) == 0 {
		return nil, errors.New("could not resolve")
	}

	return ps, nil
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
func (r *discoverer) Add(peer *peer.Peer) {
	r.cacheLock.Lock()
	r.cacheTemp[peer.Signature.PublicKey.Fingerprint().String()] = peer
	r.cacheLock.Unlock()
}

// AddPersistent allows adding permanent peer infos to be resolved.
// These peers can be overshadowed by other discoverers, but will never be gc-ed
// Mainly used for adding bootstrap nodes.
func (r *discoverer) AddPersistent(peer *peer.Peer) {
	r.cacheLock.Lock()
	r.cachePersistent[peer.Signature.PublicKey.Fingerprint().String()] = peer
	r.cacheLock.Unlock()
}
