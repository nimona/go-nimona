package resolver

import (
	"sort"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/patrickmn/go-cache"

	"nimona.io/internal/net"
	"nimona.io/internal/rand"
	"nimona.io/pkg/bloom"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

var (
	peerType               = new(peer.Peer).GetType()
	peerLookupRequestType  = new(peer.LookupRequest).GetType()
	peerLookupResponseType = new(peer.LookupResponse).GetType()
)

const (
	ErrNoPeersToAsk = errors.Error("no peers to ask")
)

//go:generate $GOBIN/mockgen -destination=../resolvermock/resolvermock_generated.go -package=resolvermock -source=resolver.go

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
		context            context.Context
		network            network.Network
		localpeer          localpeer.LocalPeer
		peerCache          *peerCache
		peerConnections    *peerConnections
		localPeerCache     *peer.Peer
		localPeerCacheLock sync.RWMutex
		// only used for initial bootstraping
		initialBootstrapPeers []*peer.Peer
		blocklist             *cache.Cache
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
		peerCache: &peerCache{
			m: sync.Map{},
		},
		peerConnections: &peerConnections{
			m: sync.Map{},
		},
		initialBootstrapPeers: []*peer.Peer{},
		blocklist:             cache.New(time.Second*5, time.Second*60),
		localPeerCacheLock:    sync.RWMutex{},
	}

	for _, opt := range opts {
		opt(r)
	}

	r.localpeer = r.network.LocalPeer()

	logger := log.FromContext(ctx).With(
		log.String("method", "resolver"),
	)

	// we are listening for all incoming object types in order to learn about
	// new peers that are talking to us so we can announce ourselves to them
	go network.HandleEnvelopeSubscription(
		r.network.Subscribe(),
		r.handleObject,
	)

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
	peersSent := make(map[crypto.PublicKey]int64)

	// send content requests to recipients
	req := &peer.LookupRequest{
		Metadata: object.Metadata{
			Owner: r.localpeer.GetPrimaryPeerKey().PublicKey(),
		},
		Nonce: rand.String(12),
		Bloom: bl,
	}
	reqObject := req.ToObject()

	peerLookupResponses := make(chan *network.Envelope)

	// listen for lookup responses
	resSub := r.network.Subscribe(
		network.FilterByObjectType(peerLookupResponseType),
		func(e *network.Envelope) bool {
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
		err := r.network.Send(
			ctx,
			reqObject,
			p,
		)
		if err != nil {
			switch err {
			case net.ErrAllAddressesBlocked,
				net.ErrNoAddresses:
				// blocklist the peer if it cannot be dialed
				r.blocklist.SetDefault(p.PublicKey().String(), p.Version)
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

				// if blocklisted skip it
				if _, blocklisted := r.blocklist.Get(
					rp.PublicKey().String(),
				); blocklisted {
					continue
				}

				// if we have already sent a high version skip
				if ver, ok := peersSent[rp.PublicKey()]; ok {
					if ver >= rp.Version {
						continue
					}
				}

				// check if the recipient matches the query
				if opt.Match(rp) {
					peers <- rp
				}

				peersSent[rp.PublicKey()] = rp.Version
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
					r.removeBlock(p)

					// if blocklisted skip it
					if _, blocklisted := r.blocklist.Get(
						p.PublicKey().String(),
					); blocklisted {
						continue
					}

					// if we have already sent a high version skip
					if ver, ok := peersSent[p.PublicKey()]; ok {
						if ver >= p.Version {
							continue
						}
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

					peersSent[p.PublicKey()] = p.Version
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
	e *network.Envelope,
) error {
	// attempt to recover correlation id from request id
	ctx := r.context

	// check if this is a peer we've received objects from in the past x minutes
	// and if not, announce ourselves to them
	lastConn := r.peerConnections.GetOrPut(e.Sender)
	if lastConn != nil && lastConn.Add(time.Minute*5).After(time.Now()) {
		r.announceSelf(e.Sender)
	}

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
	r.removeBlock(p)
}

func (r *resolver) handlePeerLookup(
	ctx context.Context,
	q *peer.LookupRequest,
	e *network.Envelope,
) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "resolver.handlePeerLookup"),
		log.String("e.sender", e.Sender.String()),
		log.Any("query.bloom", q.Bloom),
		log.Any("o.signer", e.Payload.GetSignature().Signer),
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
		Metadata: object.Metadata{
			Owner: r.localpeer.GetPrimaryPeerKey().PublicKey(),
		},
		Nonce: q.Nonce,
		Peers: cps,
	}

	p, err := r.peerCache.Get(e.Sender)
	if err != nil {
		p = &peer.Peer{
			Metadata: object.Metadata{
				Owner: e.Sender,
			},
		}
	}

	err = r.network.Send(
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
		Metadata: object.Metadata{
			Owner: r.localpeer.GetPrimaryPeerKey().PublicKey(),
		},
		Nonce: nonce,
		Bloom: r.getLocalPeer().Bloom,
	}
	o := q.ToObject()
	for _, p := range bootstrapPeers {
		logger.Debug("connecting to bootstrap", log.Strings("addresses", p.Addresses))
		err := r.network.Send(ctx, o, p)
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
		err := r.network.Send(ctx, o, p)
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
	err := r.network.Send(
		ctx,
		r.getLocalPeer().ToObject(),
		&peer.Peer{
			Metadata: object.Metadata{
				Owner: p,
			},
		},
	)
	if err != nil {
		return
	}
}

func (r *resolver) getLocalPeer() *peer.Peer {
	r.localPeerCacheLock.RLock()
	lastPeer := r.localPeerCache
	r.localPeerCacheLock.RUnlock()

	peerKey := r.localpeer.GetPrimaryPeerKey().PublicKey()
	certificates := r.localpeer.GetCertificates()
	contentHashes := r.localpeer.GetContentHashes()
	addresses := r.network.Addresses()
	relays := r.localpeer.GetRelays()

	// gather up peer key, certificates, content ids and types
	hs := []string{
		peerKey.String(),
	}
	for _, c := range contentHashes {
		hs = append(hs, c.String())
	}
	for _, c := range certificates {
		if !c.Metadata.Signature.IsEmpty() {
			hs = append(hs, c.Metadata.Signature.Signer.String())
		}
	}
	bloomSlice := bloom.New(hs...)

	if lastPeer != nil &&
		cmp.Equal(lastPeer.Addresses, addresses) &&
		cmp.Equal(lastPeer.Bloom, bloomSlice) {
		return lastPeer
	}

	localPeerCache := &peer.Peer{
		Version:      time.Now().Unix(),
		Bloom:        bloomSlice,
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

func (r *resolver) withoutOwnPeer(ps []*peer.Peer) []*peer.Peer {
	lp := r.localpeer.GetPrimaryPeerKey().PublicKey().String()
	pm := map[string]*peer.Peer{}
	for _, p := range ps {
		owner := p.Metadata.Owner
		if !owner.IsEmpty() {
			pm[owner.String()] = p
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
		// if the peer is blocklisted ignore it
		if _, blocklisted := r.blocklist.Get(
			c.peer.PublicKey().String(),
		); blocklisted {
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

func (r *resolver) removeBlock(p *peer.Peer) {
	v, blocklisted := r.blocklist.Get(p.PublicKey().String())

	if iver, ok := v.(int64); ok && blocklisted {
		if p.Version > iver {
			r.blocklist.Delete(p.PublicKey().String())
		}
	}
}
