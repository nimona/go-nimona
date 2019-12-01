package discovery

import (
	"fmt"
	"sync"

	"nimona.io/internal/rand"
	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/log"
	"nimona.io/pkg/peer"
)

//go:generate $GOBIN/mockery -name Discoverer -case underscore
//go:generate $GOBIN/mockery -name Provider -case underscore

type (
	// Provider defines the interface for a discoverer provider, eg our DHT
	Provider interface {
		Lookup(
			ctx context.Context,
			opts ...LookupOption,
		) (
			[]*peer.Peer,
			error,
		)
	}
	// Discoverer interface
	Discoverer interface {
		Add(peer *peer.Peer)
		AddProvider(provider Provider) error
		Lookup(
			ctx context.Context,
			opts ...LookupOption,
		) (
			[]*peer.Peer,
			error,
		)
	}
)

// NewDiscoverer creates a new empty discoverer with no providers
func NewDiscoverer() Discoverer {
	return &discoverer{
		providers:       sync.Map{},
		cacheTemp:       &StringPeerPeerSyncMap{},
		cachePersistent: &StringPeerPeerSyncMap{},
	}
}

// discoverer wraps multiple providers to allow resolving peer keys to peer infos
// TODO consider allowing the discoverer to accept an interface, and select
// the provider based on the input's type. This would require registering
// providers with the inputs they accept.
type discoverer struct {
	providers       sync.Map
	cacheTemp       *StringPeerPeerSyncMap
	cachePersistent *StringPeerPeerSyncMap
}

// Lookup goes through the given providers until one returns something
func (r *discoverer) Lookup(
	ctx context.Context,
	opts ...LookupOption,
) (
	[]*peer.Peer,
	error,
) {
	opt := ParseLookupOptions(opts...)

	if len(opt.Filters) == 0 {
		return nil, errors.New("missing filters")
	}

	logger := log.FromContext(ctx).With(
		log.String("method", "discovery/discoverer.Lookup"),
		log.String("opts", fmt.Sprintf("%#v", opt)),
	)

	logger.Debug("trying to lookup peers")

	ps := []*peer.Peer{}
	r.providers.Range(func(_, v interface{}) bool {
		p, ok := v.(Provider)
		if !ok {
			return true
		}
		eps, err := p.Lookup(ctx, opts...)
		if err != nil {
			logger.With(
				log.Error(err),
			).Debug("provider failed")
			return true
		}
		ps = append(ps, eps...)
		logger.With(
			log.Int("n", len(eps)),
			log.Int("total.n", len(ps)),
			log.Any("peers", ps),
		).Debug("found n peers")
		return true
	})

	// TODO move persistence into its own provider

	r.cacheTemp.Range(func(k string, p *peer.Peer) bool {
		if matchPeerWithLookupFilters(p, opt.Filters...) {
			ps = append(ps, p)
		}
		return true
	})

	r.cachePersistent.Range(func(k string, p *peer.Peer) bool {
		if matchPeerWithLookupFilters(p, opt.Filters...) {
			ps = append(ps, p)
		}
		return true
	})

	return ps, nil
}

// AddProvider to the discoverer
func (r *discoverer) AddProvider(provider Provider) error {
	r.providers.Store(rand.String(5), provider)
	return nil
}

// Add allows manually adding peer infos to be resolved.
// These peers will eventually be gc-ed.
func (r *discoverer) Add(peer *peer.Peer) {
	r.cacheTemp.Put(peer.Signature.Signer.String(), peer)
}

// AddPersistent allows adding permanent peer infos to be resolved.
// These peers can be overshadowed by other discoverers, but will never be gc-ed
// Mainly used for adding bootstrap nodes.
func (r *discoverer) AddPersistent(peer *peer.Peer) {
	r.cachePersistent.Put(peer.Signature.Signer.String(), peer)
}
