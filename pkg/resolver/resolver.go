package resolver

import (
	"sort"
	"sync"
	"time"

	"nimona.io/pkg/net"

	"nimona.io/internal/rand"
	"nimona.io/pkg/bloom"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"

	"github.com/patrickmn/go-cache"
)

var (
	peerType               = new(peer.Peer).GetType()
	peerLookupRequestType  = new(peer.LookupRequest).GetType()
	peerLookupResponseType = new(peer.LookupResponse).GetType()
)

var (
	DefaultResolver = New(
		context.Background(),
		WithEventbus(eventbus.DefaultEventbus),
		WithExchange(exchange.DefaultExchange),
		WithKeychain(keychain.DefaultKeychain),
	)
)

const (
	ErrNoPeersToAsk = errors.Error("no peers to ask")
)

type (
	Resolver interface {
		Lookup(
			ctx context.Context,
			opts ...LookupOption,
		) (<-chan *peer.Peer, error)
		Bootstrap(
			ctx context.Context,
			bootstrapPeers ...*peer.Peer,
		) error
	}
	resolver struct {
		context          context.Context
		exchange         exchange.Exchange
		eventbus         eventbus.Eventbus
		keychain         keychain.Keychain
		peerCache        *peerCache
		pinnedObjects    *pinnedObjects
		networkAddresses *networkAddresses
		relays           *relays
		// only used for initial bootstraping
		initialBootstrapPeers []*peer.Peer
		blacklist             *cache.Cache
	}
	// Option for customizing a new resolver
	Option func(*resolver)
)

// New returns a new resolver
func New(
	ctx context.Context,
	opts ...Option,
) Resolver {
	r := &resolver{
		context:  ctx,
		keychain: keychain.DefaultKeychain,
		eventbus: eventbus.DefaultEventbus,
		exchange: exchange.DefaultExchange,
		peerCache: &peerCache{
			m: sync.Map{},
		},
		pinnedObjects:         &pinnedObjects{},
		networkAddresses:      &networkAddresses{},
		relays:                &relays{},
		initialBootstrapPeers: []*peer.Peer{},
		blacklist:             cache.New(time.Second*5, time.Second*60),
	}

	for _, opt := range opts {
		opt(r)
	}

	logger := log.FromContext(ctx).With(
		log.String("method", "resolver"),
	)

	objectSub := r.exchange.Subscribe(
		exchange.FilterByObjectType(
			peerType,
			peerLookupRequestType,
			peerLookupResponseType,
		),
	)

	go exchange.HandleEnvelopeSubscription(objectSub, r.handleObject)

	// keep track of network addresses
	go func() {
		s := r.eventbus.Subscribe()
		for {
			e, err := s.Next()
			if err != nil {
				return
			}
			switch v := e.(type) {
			case eventbus.NetworkAddressAdded:
				r.networkAddresses.Put(v.Address)
			case eventbus.NetworkAddressRemoved:
				r.networkAddresses.Delete(v.Address)
			case eventbus.ObjectPinned:
				r.pinnedObjects.Put(v.Hash)
			case eventbus.ObjectUnpinned:
				r.pinnedObjects.Delete(v.Hash)
			case eventbus.PeerConnectionEstablished:
				r.announceSelf(v.PublicKey)
			case eventbus.RelayAdded:
				r.relays.Put(v.Peer)
			case eventbus.RelayRemoved:
				r.relays.Delete(v.PublicKey)
			}
		}
	}()

	// get in touch with bootstrap nodes
	go func() {
		if len(r.initialBootstrapPeers) == 0 {
			return
		}

		if err := r.Bootstrap(ctx, r.initialBootstrapPeers...); err != nil {
			logger.Error("could not bootstrap", log.Error(err))
		}

		// publish content
		if err := r.publishContentHashes(ctx); err != nil {
			logger.Error("could not publish initial content hashes", log.Error(err))
		}

		// TODO reconsider
		// subsequently try to get fresh peers every 5 minutes
		// ticker := time.NewTicker(5 * time.Minute)
		// for range ticker.C {
		// 	if _, err := r.Lookup(
		// 		context.Background(),
		// 		peer.LookupByContentType("nimona.io/peer.Peer"),
		// 	); err != nil {
		// 		logger.Error("could not refresh peers", log.Error(err))
		// 	}
		// 	if err := r.publishContentHashes(ctx); err != nil {
		// 		logger.Error("could not refresh content hashes", log.Error(err))
		// 	}
		// }
	}()

	return r
}

func Lookup(
	ctx context.Context,
	opts ...LookupOption,
) (<-chan *peer.Peer, error) {
	return DefaultResolver.Lookup(ctx, opts...)
}

// Lookup finds and returns peer infos from a fingerprint
func (r *resolver) Lookup(
	ctx context.Context,
	opts ...LookupOption,
) (<-chan *peer.Peer, error) {
	opt := ParseLookupOptions(opts...)

	logger := log.FromContext(ctx).With(
		log.String("method", "resolver.Lookup"),
	)
	logger.Debug("looking up")

	bl := bloom.New(opt.Lookups...)

	// create channel to keep peers we find
	peers := make(chan *peer.Peer, 100)

	// send content requests to recipients
	req := &peer.LookupRequest{
		Nonce: rand.String(12),
		Bloom: bl,
	}
	reqObject := req.ToObject()

	peerLookupResponses := make(chan *exchange.Envelope)

	// listen for lookup responses
	resSub := r.exchange.Subscribe(
		exchange.FilterByObjectType(peerLookupResponseType),
		func(e *exchange.Envelope) bool {
			v := e.Payload.Get("nonce:s")
			rn, ok := v.(string)
			return ok && rn == req.Nonce
		},
	)
	go func() {
		defer close(peerLookupResponses)
		for {
			e, err := resSub.Next()
			if err != nil {
				break
			}
			peerLookupResponses <- e
		}
	}()

	// create channel for the peers we need to ask
	initialRecipients := make(chan *peer.Peer, 100)

	queryPeer := func(p *peer.Peer) {
		err := r.exchange.Send(
			ctx,
			reqObject,
			p,
		)
		if err != nil {
			switch err {
			case net.ErrAllAddressesBlacklisted,
				net.ErrNoAddresses:
				// blacklist the peer if it cannot be dialed
				r.blacklist.SetDefault(p.PublicKey().String(), p.Version)
			}
			logger.Debug("could send request to peer", log.Error(err))
			return
		}
		logger.Debug("asked peer", log.String("peer", p.PublicKey().String()))
	}

	go func() {
		// keep a record of who responded
		recipientsResponded := map[crypto.PublicKey]bool{}
		// just in case timeout
		// TODO maybe figure out if the ctx as a timeout before adding one
		timeout := time.NewTimer(time.Second * 10)
		defer close(initialRecipients)
		defer close(peers)
		defer resSub.Cancel()
		for {
			select {
			case <-ctx.Done():
				logger.Debug("ctx done, giving up")
				return
			case <-timeout.C:
				logger.Debug("timeout done, giving up")
				return
			case rp, ok := <-initialRecipients:
				if !ok {
					initialRecipients = nil
					continue
				}

				// if blacklisted skip it
				if _, blacklisted := r.blacklist.Get(
					rp.PublicKey().String(),
				); blacklisted {
					continue
				}

				// check if the recipient matches the query
				if opt.Match(rp) {
					peers <- rp
				}
				// mark peer as asked
				recipientsResponded[rp.PublicKey()] = false
				// ask recipient
				queryPeer(rp)
			case e, ok := <-peerLookupResponses:
				if !ok {
					peerLookupResponses = nil
					continue
				}
				res := &peer.LookupResponse{}
				if err := res.FromObject(e.Payload); err != nil {
					continue
				}
				// mark sender as responded
				recipientsResponded[e.Sender] = true
				for _, p := range res.Peers {
					// add peers to our peerstore
					// TODO pin this in the cache
					r.peerCache.Put(p)
					r.removeBlacklist(p)

					// if blacklisted skip it
					if _, blacklisted := r.blacklist.Get(
						p.PublicKey().String(),
					); blacklisted {
						continue
					}
					// if the peer matches the query, add it to our results
					if opt.Match(p) {
						peers <- p
					}
					// check if we've already asked this peer
					if _, asked := recipientsResponded[p.PublicKey()]; asked {
						// if so, move on
						continue
					}
					// else mark peer as asked
					recipientsResponded[p.PublicKey()] = false
					// and ask them
					queryPeer(p)
				}
				allDone := true
				for _, answered := range recipientsResponded {
					if !answered {
						allDone = false
					}
				}
				if allDone {
					return
				}
			}
		}
	}()

	cps := r.getClosest(bl, 10)
	cps = r.withoutOwnPeer(cps)
	for _, p := range cps {
		initialRecipients <- p
	}

	return peers, nil
}

func (r *resolver) handleObject(
	e *exchange.Envelope,
) error {
	// attempt to recover correlation id from request id
	ctx := r.context

	// handle payload
	o := e.Payload
	switch o.GetType() {
	case peerType:
		v := &peer.Peer{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeer(ctx, v)
	case peerLookupRequestType:
		v := &peer.LookupRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerLookup(ctx, v, e)
	case peerLookupResponseType:
		v := &peer.LookupResponse{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		for _, p := range v.Peers {
			r.handlePeer(ctx, p)
		}
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
	r.peerCache.Put(p)
	r.removeBlacklist(p)
}

func (r *resolver) handlePeerLookup(
	ctx context.Context,
	q *peer.LookupRequest,
	e *exchange.Envelope,
) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "resolver.handlePeerLookup"),
		log.String("e.sender", e.Sender.String()),
		log.Any("query.bloom", q.Bloom),
		log.Any("o.signer", e.Payload.GetSignatures()),
	)

	logger.Debug("handling peer lookup")

	ps := r.getClosest(q.Bloom, 30)
	// keep only peers which have addresses or relays
	cps := []*peer.Peer{}
	for _, p := range ps {
		if len(p.Addresses) != 0 || len(p.Relays) != 0 {
			cps = append(cps, p)
		}
	}
	// TODO why are we adding ourselves?
	cps = append(cps, r.getLocalPeer())
	cps = peer.Unique(cps)

	ctx = context.New(
		context.WithParent(ctx),
	)

	res := &peer.LookupResponse{
		Nonce: q.Nonce,
		Peers: cps,
	}

	p, err := r.peerCache.Get(e.Sender)
	if err != nil {
		p = &peer.Peer{
			Owners: []crypto.PublicKey{
				e.Sender,
			},
		}
	}

	err = r.exchange.Send(
		ctx,
		res.ToObject(),
		p,
	)
	if err != nil {
		logger.Debug("could not send lookup response",
			log.Error(err),
		)
	}
	logger.With(
		log.Int("n", len(cps)),
	).Debug("handling done, sent n peers")
}

func (r *resolver) Bootstrap(
	ctx context.Context,
	bootstrapPeers ...*peer.Peer,
) error {
	logger := log.FromContext(ctx)
	nonce := rand.String(6)
	q := &peer.LookupRequest{
		Nonce: nonce,
		Bloom: r.getLocalPeer().Bloom,
	}
	o := q.ToObject()
	for _, p := range bootstrapPeers {
		logger.Debug("connecting to bootstrap", log.Strings("addresses", p.Addresses))
		err := r.exchange.Send(ctx, o, p)
		if err != nil {
			logger.Debug("could not send request to bootstrap", log.Error(err))
		}
	}
	return nil
}

func (r *resolver) publishContentHashes(
	ctx context.Context,
) error {
	logger := log.FromContext(ctx).With(
		log.String("method", "resolver.publishContentHashes"),
	)
	cb := r.getLocalPeer()
	ps := r.getClosest(cb.Bloom, 10)
	if len(ps) == 0 {
		logger.Debug("couldn't find peers to tell")
		return errors.New("no peers to tell")
	}

	logger.With(
		log.Int("n", len(ps)),
		log.Any("bloom", cb.Bloom),
	).Debug("trying to tell n peers")

	o := cb.ToObject()
	for _, p := range ps {
		err := r.exchange.Send(ctx, o, p)
		if err != nil {
			logger.Debug("could not send request", log.Error(err))
		}
	}
	return nil
}

func (r *resolver) announceSelf(p crypto.PublicKey) {
	ctx := context.New(
		context.WithTimeout(time.Second * 3),
	)
	err := r.exchange.Send(
		ctx,
		r.getLocalPeer().ToObject(),
		&peer.Peer{
			Owners: []crypto.PublicKey{p},
		},
	)
	if err != nil {
		return
	}
}

func (r *resolver) getLocalPeer() *peer.Peer {
	k := r.keychain.GetPrimaryPeerKey()
	pk := k.PublicKey()
	cs := r.keychain.GetCertificates(pk)

	// gather up peer key, certificates, content ids and types
	hs := []string{
		pk.String(),
	}

	for _, c := range r.pinnedObjects.List() {
		hs = append(hs, c.String())
	}

	for _, c := range cs {
		hs = append(hs, c.Subject.String())
	}

	// TODO cache peer info and reuse
	pi := &peer.Peer{
		Version:   time.Now().UTC().Unix(),
		Bloom:     bloom.New(hs...),
		Addresses: r.networkAddresses.List(),
		Relays:    r.relays.List(),
		Certificates: r.keychain.GetCertificates(
			r.keychain.GetPrimaryPeerKey().PublicKey(),
		),
		Owners: r.keychain.ListPublicKeys(keychain.PeerKey),
	}

	o := pi.ToObject()
	sig, err := object.NewSignature(k, o)
	if err != nil {
		panic(err)
	}

	pi.Signatures = append(pi.Signatures, sig)

	return pi
}

func (r *resolver) withoutOwnPeer(ps []*peer.Peer) []*peer.Peer {
	lp := r.keychain.GetPrimaryPeerKey().PublicKey().String()
	pm := map[string]*peer.Peer{}
	for _, p := range ps {
		for _, s := range p.Owners {
			pm[s.String()] = p
		}
	}
	nps := []*peer.Peer{}
	for f, p := range pm {
		if f == lp {
			continue
		}
		nps = append(nps, p)
	}
	return nps
}

// getClosest returns the closest peers to the bloom filter from the cache
func (r *resolver) getClosest(q bloom.Bloom, n int) []*peer.Peer {
	type kv struct {
		bloomIntersection int
		peer              *peer.Peer
	}

	rs := []kv{}
	for _, p := range r.peerCache.List() {
		rs = append(rs, kv{
			bloomIntersection: intersectionCount(
				q.Bloom(),
				p.Bloom,
			),
			peer: p,
		})
	}

	if len(rs) == 0 {
		return []*peer.Peer{}
	}

	sort.Slice(rs, func(i, j int) bool {
		return rs[i].bloomIntersection < rs[j].bloomIntersection
	})

	fs := []*peer.Peer{}
	for i, c := range rs {
		// if the peer is blacklisted ignore it
		if _, blacklisted := r.blacklist.Get(
			c.peer.PublicKey().String(),
		); blacklisted {
			continue
		}
		fs = append(fs, c.peer)
		if i > n {
			break
		}
	}

	return fs
}

func intersectionCount(a, b []int64) int {
	m := make(map[int64]uint64)
	for _, k := range a {
		m[k] |= (1 << 0)
	}
	for _, k := range b {
		m[k] |= (1 << 1)
	}

	i := 0
	for _, v := range m {
		a := v&(1<<0) != 0
		b := v&(1<<1) != 0
		if a && b {
			i++
		}
	}

	return i
}

func (r *resolver) removeBlacklist(p *peer.Peer) {
	v, blacklisted := r.blacklist.Get(p.PublicKey().String())

	if iver, ok := v.(int64); ok && blacklisted {
		if p.Version > iver {
			r.blacklist.Delete(p.PublicKey().String())
		}
	}
}

func Bootstrap(
	ctx context.Context,
	bootstrapPeers ...*peer.Peer,
) error {
	return DefaultResolver.Bootstrap(ctx, bootstrapPeers...)
}
