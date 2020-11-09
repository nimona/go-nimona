package resolver

import (
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/patrickmn/go-cache"

	"nimona.io/internal/rand"
	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/hyperspace"
	"nimona.io/pkg/hyperspace/peerstore"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

var (
	peerType               = new(peer.Peer).Type()
	peerLookupResponseType = new(peer.LookupResponse).Type()

	peerCacheTTL = 1 * time.Minute
)

const (
	ErrNoPeersToAsk = errors.Error("no peers to ask")
)

//go:generate mockgen -destination=../resolvermock/resolvermock_generated.go -package=resolvermock -source=resolver.go

type (
	Resolver interface {
		Lookup(
			ctx context.Context,
			opts ...LookupOption,
		) ([]*peer.Peer, error)
	}
	resolver struct {
		context            context.Context
		network            network.Network
		localpeer          localpeer.LocalPeer
		peerCache          *peerstore.PeerCache
		peerConnections    *peerConnections
		localPeerCache     *peer.Peer
		localPeerCacheLock sync.RWMutex
		bootstrapPeers     []*peer.Peer
		blocklist          *cache.Cache
	}
	// Option for customizing a new resolver
	Option func(*resolver)
)

// New returns a new resolver
func New(
	ctx context.Context,
	netw network.Network,
	opts ...Option,
) Resolver {
	r := &resolver{
		context: ctx,
		network: netw,
		peerCache: peerstore.NewPeerCache(
			time.Minute,
			"nimona_hyperspace_resolver",
		),
		peerConnections: &peerConnections{
			m: sync.Map{},
		},
		localPeerCacheLock: sync.RWMutex{},
		bootstrapPeers:     []*peer.Peer{},
		blocklist:          cache.New(time.Second*5, time.Second*60),
	}

	for _, opt := range opts {
		opt(r)
	}

	r.localpeer = r.network.LocalPeer()

	// we are listening for all incoming object types in order to learn about
	// new peers that are talking to us so we can announce ourselves to them
	go network.HandleEnvelopeSubscription(
		r.network.Subscribe(),
		r.handleObject,
	)

	for _, p := range r.bootstrapPeers {
		r.peerCache.Put(p, 0)
	}

	go func() {
		r.announceSelf()
		announceTimer := time.NewTicker(30 * time.Second)
		for range announceTimer.C {
			r.announceSelf()
		}
	}()

	return r
}

// Lookup finds and returns peer infos from a fingerprint
// TODO consider returning peers synchronously
func (r *resolver) Lookup(
	ctx context.Context,
	opts ...LookupOption,
) ([]*peer.Peer, error) {
	if len(r.bootstrapPeers) == 0 {
		return nil, errors.New("no peers to ask")
	}

	logger := log.FromContext(ctx).With(
		log.String("method", "resolver.Lookup"),
	)
	logger.Debug("looking up")

	opt := ParseLookupOptions(opts...)
	bl := hyperspace.New(opt.Lookups...)

	// send content requests to recipients
	req := &peer.LookupRequest{
		Metadata: object.Metadata{
			Owner: r.localpeer.GetPrimaryPeerKey().PublicKey(),
		},
		Nonce:       rand.String(12),
		QueryVector: bl,
	}
	reqObject := req.ToObject()

	// listen for lookup responses
	resSub := r.network.Subscribe(
		network.FilterByObjectType(peerLookupResponseType),
		func(e *network.Envelope) bool {
			v := e.Payload.Data["nonce:s"]
			rn, ok := v.(string)
			return ok && rn == req.Nonce
		},
	)

	go func() {
		for _, bp := range r.bootstrapPeers {
			err := r.network.Send(
				ctx,
				reqObject,
				bp,
			)
			if err != nil {
				logger.Debug("could send request to peer", log.Error(err))
				continue
			}
			logger.Debug(
				"asked peer",
				log.String("peer", bp.PublicKey().String()),
			)
		}
	}()

	// create channel to keep peers we find
	peers := []*peer.Peer{}
	done := make(chan struct{})

	go func() {
		for {
			e, err := resSub.Next()
			if err != nil {
				break
			}
			r := &peer.LookupResponse{}
			if err := r.FromObject(e.Payload); err != nil {
				continue
			}
			// TODO verify peer?
			peers = append(peers, r.Peers...)
			close(done)
			break
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		return peers, nil
	}
}

func (r *resolver) handleObject(
	e *network.Envelope,
) error {
	// attempt to recover correlation id from request id
	ctx := r.context

	// handle payload
	o := e.Payload
	if o.Type == peerType {
		v := &peer.Peer{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeer(ctx, v)
	}
	return nil
}

func (r *resolver) handlePeer(
	ctx context.Context,
	p *peer.Peer,
) {
	logger := log.FromContext(ctx).With(
		log.String("method", "resolver.handlePeer"),
		log.String("peer.publicKey", p.PublicKey().String()),
		log.Strings("peer.addresses", p.Addresses),
	)
	logger.Debug("adding peer to cache")
	r.peerCache.Put(p, peerCacheTTL)
}

func (r *resolver) announceSelf() {
	ctx := context.New(
		context.WithParent(r.context),
	)
	logger := log.FromContext(ctx).With(
		log.String("method", "resolver.announceSelf"),
	)
	n := 0
	for _, p := range r.bootstrapPeers {
		if err := r.network.Send(
			context.New(
				context.WithParent(ctx),
				context.WithTimeout(time.Second*3),
			),
			r.getLocalPeer().ToObject(),
			p,
		); err != nil {
			logger.Error(
				"error announcing self to bootstrap",
				log.String("peer", p.PublicKey().String()),
				log.Error(err),
			)
			continue
		}
		n++
	}
	logger.Info("announced self to bootstrap peers", log.Int("n", n))
}

func (r *resolver) getLocalPeer() *peer.Peer {
	r.localPeerCacheLock.RLock()
	lastPeer := r.localPeerCache
	r.localPeerCacheLock.RUnlock()

	peerKey := r.localpeer.GetPrimaryPeerKey().PublicKey()
	certificates := r.localpeer.GetCertificates()
	contentHashes := r.localpeer.GetContentHashes()
	contentTypes := r.localpeer.GetContentTypes()
	addresses := r.localpeer.GetAddresses()
	relays := r.localpeer.GetRelays()

	// gather up peer key, certificates, content ids and types
	hs := contentTypes
	hs = append(hs, peerKey.String())
	for _, c := range contentHashes {
		hs = append(hs, c.String())
	}
	for _, c := range certificates {
		if !c.Metadata.Signature.IsEmpty() {
			hs = append(hs, c.Metadata.Signature.Signer.String())
		}
	}
	vec := hyperspace.New(hs...)

	if lastPeer != nil &&
		cmp.Equal(lastPeer.Addresses, addresses) &&
		cmp.Equal(lastPeer.QueryVector, vec) {
		return lastPeer
	}

	localPeerCache := &peer.Peer{
		Version:      time.Now().Unix(),
		QueryVector:  vec,
		Addresses:    addresses,
		Relays:       relays,
		Certificates: certificates,
		Metadata: object.Metadata{
			Owner: peerKey,
		},
	}

	r.localPeerCacheLock.Lock()
	r.localPeerCache = localPeerCache
	r.localPeerCacheLock.Unlock()

	return localPeerCache
}
