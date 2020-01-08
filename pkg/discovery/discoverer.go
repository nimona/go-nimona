package discovery

import (
	"fmt"
	"sync"

	"nimona.io/internal/rand"
	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/log"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
)

//go:generate $GOBIN/mockery -name PeerStorer -case underscore
//go:generate $GOBIN/mockery -name Discoverer -case underscore

type (
	// Discoverer defines the interface for a addressBook provider, eg our DHT
	Discoverer interface {
		Lookup(ctx context.Context, opts ...peer.LookupOption) ([]*peer.Peer, error)
	}
	// PeerStorer interface
	PeerStorer interface {
		Add(*peer.Peer, bool)
		AddDiscoverer(Discoverer) error
		Lookup(context.Context, ...peer.LookupOption) ([]*peer.Peer, error)
	}
)

// NewPeerStorer creates a new empty addressBook with no providers
func NewPeerStorer(store *sqlobjectstore.Store) PeerStorer {
	return &addressBook{
		providers: sync.Map{},
		store:     store,
	}
}

// addressBook is
type addressBook struct {
	providers sync.Map
	store     *sqlobjectstore.Store
}

// Lookup goes through the given providers until one returns something
func (r *addressBook) Lookup(
	ctx context.Context,
	opts ...peer.LookupOption,
) ([]*peer.Peer, error) {
	opt := peer.ParseLookupOptions(opts...)

	if len(opt.Filters) == 0 {
		return nil, errors.New("missing filters")
	}

	logger := log.FromContext(ctx).With(
		log.String("method", "discovery/addressBook.Lookup"),
		log.String("opts", fmt.Sprintf("%#v", opt)),
	)

	logger.Debug("trying to lookup peers")

	// find all peer objects
	// TODO replace with sqlpeerstore
	os, err := r.store.Filter(sqlobjectstore.FilterByObjectType("nimona.io/peer.Peer"))
	if err != nil {
		return nil, err
	}

	// something to hold our results
	ps := []*peer.Peer{}

	// go through the peers objects, and if they match the lookup options
	// add them to the results
	for _, o := range os {
		p := &peer.Peer{}
		if err := p.FromObject(o); err != nil {
			continue
		}
		if opt.Match(p) {
			ps = append(ps, p)
		}
	}

	// if we have found some results, let's just return
	// TODO I don't really like this
	if len(ps) > 0 {
		return ps, nil
	}

	// if no results have been found but only local results were requests return
	if opt.Local {
		return ps, nil
	}

	// else, go through the discoveres and try to find some results
	// TODO once we have more than one discoverer we should run them in parallel
	r.providers.Range(func(_, v interface{}) bool {
		p, ok := v.(Discoverer)
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

	return ps, nil
}

// AddDiscoverer to the addressBook
func (r *addressBook) AddDiscoverer(provider Discoverer) error {
	r.providers.Store(rand.String(5), provider)
	return nil
}

// Add allows manually adding peer infos to be resolved.
// These peers will eventually be gc-ed unless pinned.
// WARNING: Only bootstrap peers should be pinned. Probably.
func (r *addressBook) Add(peer *peer.Peer, pin bool) {
	opts := []sqlobjectstore.Option{}
	if pin {
		opts = append(opts, sqlobjectstore.WithTTL(0))
	} else {
		opts = append(opts, sqlobjectstore.WithTTL(60))
	}
	r.store.Put(peer.ToObject(), opts...)
}
