package hyperspace

import (
	"sort"
	"time"

	"nimona.io/internal/rand"
	"nimona.io/pkg/bloom"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/eventbus"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/keychain"
	"nimona.io/pkg/log"
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

type (
	// Discoverer hyperspace
	Discoverer struct {
		context          context.Context
		peerstore        discovery.PeerStorer
		exchange         exchange.Exchange
		eventbus         eventbus.Eventbus
		keychain         keychain.Keychain
		pinnedObjects    *pinnedObjects
		networkAddresses *networkAddresses
		relays           *relays
	}
)

// New returns a new discoverer
func New(
	ctx context.Context,
	ps discovery.PeerStorer,
	kc keychain.Keychain,
	eb eventbus.Eventbus,
	exc exchange.Exchange,
	bootstrapPeers []*peer.Peer,
) (*Discoverer, error) {
	r := &Discoverer{
		context:          ctx,
		peerstore:        ps,
		eventbus:         eb,
		keychain:         kc,
		exchange:         exc,
		pinnedObjects:    &pinnedObjects{},
		networkAddresses: &networkAddresses{},
		relays:           &relays{},
	}

	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/Discoverer"),
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
				r.relays.Put(v.PublicKey)
			case eventbus.RelayRemoved:
				r.relays.Delete(v.PublicKey)
			}
		}
	}()

	// get in touch with bootstrap nodes
	go func() {
		if err := r.bootstrap(ctx, bootstrapPeers); err != nil {
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

	return r, nil
}

// Lookup finds and returns peer infos from a fingerprint
func (r *Discoverer) Lookup(
	ctx context.Context,
	opts ...peer.LookupOption,
) (<-chan *peer.Peer, error) {
	opt := peer.ParseLookupOptions(opts...)

	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.Lookup"),
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
	initialRecipients := make(chan crypto.PublicKey, 100)

	queryPeer := func(pk crypto.PublicKey) {
		err := r.exchange.Send(
			ctx,
			reqObject,
			peer.LookupByOwner(pk),
			exchange.WithLocalDiscoveryOnly(),
			exchange.WithAsync(),
		)
		if err != nil {
			logger.Debug("could send request to peer", log.Error(err))
		}
		logger.Debug("asked peer", log.String("peer", pk.String()))
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
			case r, ok := <-initialRecipients:
				if !ok {
					initialRecipients = nil
					continue
				}
				// mark peer as asked
				recipientsResponded[r] = false
				// ask recipient
				queryPeer(r)
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
					r.peerstore.Add(p, false)
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
					queryPeer(p.PublicKey())
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

	aps, err := r.peerstore.Lookup(ctx, peer.LookupOnlyLocal())
	if err != nil {
		logger.Error("error getting all peers", log.Error(err))
		return nil, err
	}

	pps := []*peer.Peer{}
	for p := range aps {
		pps = append(pps, p)
	}
	cps := getClosest(pps, bl)
	cps = r.withoutOwnPeer(cps)
	for _, p := range cps {
		initialRecipients <- p.PublicKey()
	}

	return peers, nil
}

func (r *Discoverer) handleObject(
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

func (r *Discoverer) handlePeer(
	ctx context.Context,
	p *peer.Peer,
) {
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.handlePeer"),
		log.String("peer.publicKey", p.PublicKey().String()),
		log.Strings("peer.addresses", p.Addresses),
	)
	logger.Debug("adding peer to store")
	r.peerstore.Add(p, false)
}

func (r *Discoverer) handlePeerLookup(
	ctx context.Context,
	q *peer.LookupRequest,
	e *exchange.Envelope,
) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.handlePeerLookup"),
		log.String("e.sender", e.Sender.String()),
		log.Any("query.bloom", q.Bloom),
		log.Any("o.signer", e.Payload.GetSignatures()),
	)

	logger.Debug("handling peer lookup")

	aps, err := r.peerstore.Lookup(ctx, peer.LookupOnlyLocal())
	if err != nil {
		return
	}
	pps := []*peer.Peer{}
	for p := range aps {
		pps = append(pps, p)
	}
	cps := getClosest(pps, q.Bloom)
	cps = append(cps, r.getLocalPeer())
	cps = peer.Unique(cps)

	ctx = context.New(
		context.WithParent(ctx),
	)

	res := &peer.LookupResponse{
		Nonce: q.Nonce,
		Peers: cps,
	}

	err = r.exchange.Send(
		ctx,
		res.ToObject(),
		peer.LookupByOwner(e.Sender),
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithAsync(),
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

func (r *Discoverer) bootstrap(
	ctx context.Context,
	bootstrapPeers []*peer.Peer,
) error {
	logger := log.FromContext(ctx)
	opts := []exchange.SendOption{
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithAsync(),
	}
	nonce := rand.String(6)
	q := &peer.LookupRequest{
		Nonce: nonce,
		Bloom: r.getLocalPeer().Bloom,
	}
	o := q.ToObject()
	for _, p := range bootstrapPeers {
		logger.Debug("connecting to bootstrap", log.Strings("addresses", p.Addresses))
		err := r.exchange.SendToPeer(ctx, o, p, opts...)
		if err != nil {
			logger.Debug("could not send request to bootstrap", log.Error(err))
		}
	}
	return nil
}

func (r *Discoverer) publishContentHashes(
	ctx context.Context,
) error {
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/Discoverer.publishContentHashes"),
	)
	cb := r.getLocalPeer()
	aps, err := r.peerstore.Lookup(ctx, peer.LookupOnlyLocal())
	if err != nil {
		return err
	}
	pps := []*peer.Peer{}
	for p := range aps {
		pps = append(pps, p)
	}
	cps := getClosest(pps, cb.Bloom)
	fs := []crypto.PublicKey{}
	for _, c := range cps {
		fs = append(fs, c.Owners...)
	}
	if len(fs) == 0 {
		logger.Debug("couldn't find peers to tell")
		return errors.New("no peers to tell")
	}

	logger.With(
		log.Int("n", len(fs)),
		log.Any("bloom", cb.Bloom),
	).Debug("trying to tell n peers")

	opts := []exchange.SendOption{
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithAsync(),
	}

	o := cb.ToObject()
	for _, f := range fs {
		err := r.exchange.Send(ctx, o, peer.LookupByOwner(f), opts...)
		if err != nil {
			logger.Debug("could not send request", log.Error(err))
		}
	}
	return nil
}

func (r *Discoverer) announceSelf(p crypto.PublicKey) {
	ctx := context.New(
		context.WithTimeout(time.Second * 3),
	)
	err := r.exchange.SendToPeer(
		ctx,
		r.getLocalPeer().ToObject(),
		&peer.Peer{
			Version: -1,
			Owners:  []crypto.PublicKey{p},
		},
		exchange.WithAsync(),
	)
	if err != nil {
		return
	}
}

func (r *Discoverer) getLocalPeer() *peer.Peer {
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

func (r *Discoverer) withoutOwnPeer(ps []*peer.Peer) []*peer.Peer {
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

// getClosest returns peers that closest resemble the query
func getClosest(ps []*peer.Peer, q bloom.Bloom) []*peer.Peer {
	type kv struct {
		bloomIntersection int
		peer              *peer.Peer
	}

	r := []kv{}
	for _, p := range ps {
		r = append(r, kv{
			bloomIntersection: intersectionCount(
				q.Bloom(),
				p.Bloom,
			),
			peer: p,
		})
	}

	sort.Slice(r, func(i, j int) bool {
		return r[i].bloomIntersection < r[j].bloomIntersection
	})

	fs := []*peer.Peer{}
	for i, c := range r {
		fs = append(fs, c.peer)
		if i > 10 { // TODO make limit configurable
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
