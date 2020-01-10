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
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/log"
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
		context   context.Context
		peerstore discovery.PeerStorer
		// store     *Store
		exchange exchange.Exchange
		local    *peer.LocalPeer
	}
)

// NewDiscoverer returns a new hyperspace discoverer
func NewDiscoverer(
	ctx context.Context,
	ps discovery.PeerStorer,
	exc exchange.Exchange,
	local *peer.LocalPeer,
	bootstrapAddresses []string,
) (*Discoverer, error) {
	r := &Discoverer{
		context:   ctx,
		peerstore: ps,
		// store:     NewStore(),
		local:    local,
		exchange: exc,
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

	// get in touch with bootstrap nodes
	if err := r.bootstrap(ctx, bootstrapAddresses); err != nil {
		logger.Error("could not bootstrap", log.Error(err))
	}

	// publish content
	if err := r.publishContentHashes(ctx); err != nil {
		logger.Error("could not publish initial content hashes", log.Error(err))
	}

	return r, nil
}

// Lookup finds and returns peer infos from a fingerprint
func (r *Discoverer) Lookup(
	ctx context.Context,
	opts ...peer.LookupOption,
) (
	[]*peer.Peer,
	error,
) {
	opt := peer.ParseLookupOptions(opts...)

	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.Lookup"),
	)
	logger.Debug("looking up")

	eps, err := r.lookup(
		ctx,
		bloom.New(opt.Lookups...),
		opt,
	)
	if err != nil {
		return nil, err
	}

	ps := r.withoutOwnPeer(eps)

	return ps, nil
}

func (r *Discoverer) handleObject(
	e *exchange.Envelope,
) error {
	// attempt to recover correlation id from request id
	ctx := r.context

	// logger := log.FromContext(ctx).With(
	// 	log.String("method", "hyperspace/resolver.handleObject"),
	// )

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
	// logger.Debug("added peer to store")
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
	)

	logger.Debug("handling peer lookup")

	aps, err := r.peerstore.Lookup(ctx, peer.LookupOnlyLocal())
	if err != nil {
		return
	}
	cps := getClosest(aps, q.Bloom)
	res := &peer.LookupResponse{
		Nonce: q.Nonce,
		Peers: []*peer.Peer{},
	}

	ctx = context.New(
		context.WithParent(ctx),
	)

	for _, p := range cps {
		logger.Debug("responding for content hash bloom",
			log.Any("bloom", p.Bloom),
			log.String("peer", p.Signature.Signer.String()),
		)
		err := r.exchange.Send(
			ctx,
			p.ToObject(),
			peer.LookupByKey(e.Sender),
			exchange.WithLocalDiscoveryOnly(),
			exchange.WithAsync(),
		)
		if err != nil {
			logger.Debug("could not send peer",
				log.Error(err),
			)
			continue
		}
		res.Peers = append(res.Peers, p)
	}
	err = r.exchange.Send(
		ctx,
		res.ToObject(),
		peer.LookupByKey(e.Sender),
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

// lookup does a network lookup given a query
func (r *Discoverer) lookup(
	ctx context.Context,
	bloom bloom.Bloom,
	matcher *peer.LookupOptions,
) ([]*peer.Peer, error) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.LookupPeer"),
		log.Any("bloom", bloom),
	)
	logger.Debug("looking up peer")

	// create channel to mark all requests are completed
	done := make(chan struct{})
	// create channel to keep peers we find
	peers := make(chan *peer.Peer)
	// create channel for the peers we need to ask
	recipients := make(chan crypto.PublicKey)
	// and for the ones that responded
	recipientsResponded := make(chan crypto.PublicKey)

	// // allows us to match peers with peerRequest
	// peerMatchesRequest := func(p *peer.Peer) bool {
	// 	return matcher.Match(p)
	// }

	// send content requests to recipients
	req := &peer.LookupRequest{
		Nonce: rand.String(12),
		Bloom: bloom,
	}
	reqObject := req.ToObject()
	go func() {
		// keep a record of the peers we already asked
		// also mark our own peer as done from the beginning
		recipientsAsked := map[string]bool{
			r.local.GetPeerPublicKey().String(): true,
		}
		for {
			select {
			// go through the peers we need to ask
			case recipient := <-recipients:
				// check if we've already asked them
				if _, asked := recipientsAsked[recipient.String()]; asked {
					continue
				}
				// else mark them as already been asked
				recipientsAsked[recipient.String()] = false
				// and finally ask them
				err := r.exchange.Send(
					ctx,
					reqObject,
					peer.LookupByKey(recipient),
					exchange.WithLocalDiscoveryOnly(),
					exchange.WithAsync(),
				)
				if err != nil {
					logger.Debug("could send request to peer", log.Error(err))
				}
				logger.Debug("asked peer", log.String("peer", recipient.String()))
			case recipient := <-recipientsResponded:
				recipientsAsked[recipient.String()] = true
				allDone := true
				for _, ok := range recipientsAsked {
					if !ok {
						allDone = false
						break
					}
				}
				if allDone {
					done <- struct{}{}
				}
			}
		}
	}()

	// listen for lookup responses
	resSub := r.exchange.Subscribe(
		exchange.FilterByObjectType(peerLookupResponseType),
		func(e *exchange.Envelope) bool {
			v := e.Payload.Get("nonce:s")
			rn, ok := v.(string)
			return ok && rn == req.Nonce
		},
	)
	defer resSub.Cancel()
	go func() {
		for {
			e, err := resSub.Next()
			if err != nil {
				break
			}
			res := &peer.LookupResponse{}
			if err := res.FromObject(e.Payload); err != nil {
				break
			}
			// logger.Debug("_____ !!!! got resp", log.Int("n", len(res.Peers)))
			for _, p := range res.Peers {
				// logger.Debug("_____ !!!! got resp", log.String("key", r.PublicKey().String()))
				if matcher.Match(p) {
					peers <- p
					continue
				}
				recipients <- p.PublicKey()
			}
			recipientsResponded <- e.Sender
		}
	}()

	// find and ask "closest" peers
	go func() {
		aps, err := r.peerstore.Lookup(ctx, peer.LookupOnlyLocal())
		if err != nil {
			logger.Error("error getting all peers", log.Error(err))
			return
		}
		cps := getClosest(aps, bloom)
		cps = r.withoutOwnPeer(cps)
		for _, p := range cps {
			if matcher.Match(p) {
				peers <- p
				continue
			}
			recipients <- p.PublicKey()
		}
	}()

	// gather all peers until something happens
	peersList := []*peer.Peer{}
	t := time.NewTimer(time.Second * 5)
loop:
	for {
		select {
		case peer := <-peers:
			peersList = append(peersList, peer)
			break loop
		case <-done:
			logger.Debug("done")
			break loop
		case <-t.C:
			logger.Debug("timer done, giving up")
			break loop
		case <-ctx.Done():
			logger.Debug("ctx done, giving up")
			break loop
		}
	}

	logger.Debug("done, found n peers", log.Int("n", len(peersList)))
	return peersList, nil
}

func (r *Discoverer) bootstrap(
	ctx context.Context,
	bootstrapAddresses []string,
) error {
	logger := log.FromContext(ctx)
	opts := []exchange.Option{
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithAsync(),
	}
	nonce := rand.String(6)
	q := &peer.LookupRequest{
		Nonce: nonce,
		Bloom: r.local.GetSignedPeer().Bloom,
	}
	o := q.ToObject()
	for _, addr := range bootstrapAddresses {
		logger.Debug("connecting to bootstrap", log.String("address", addr))
		err := r.exchange.SendToAddress(ctx, o, addr, opts...)
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
	cb := r.local.GetSignedPeer()
	aps, err := r.peerstore.Lookup(ctx, peer.LookupOnlyLocal())
	if err != nil {
		return err
	}
	cps := getClosest(aps, cb.Bloom)
	fs := []crypto.PublicKey{}
	for _, c := range cps {
		if c.Signature != nil {
			fs = append(fs, c.Signature.Signer)
		}
	}
	if len(fs) == 0 {
		logger.Debug("couldn't find peers to tell")
		return errors.New("no peers to tell")
	}

	logger.With(
		log.Int("n", len(fs)),
		log.Any("bloom", cb.Bloom),
	).Debug("trying to tell n peers")

	opts := []exchange.Option{
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithAsync(),
	}

	o := cb.ToObject()
	if err := crypto.Sign(o, r.local.GetPeerPrivateKey()); err != nil {
		logger.With(
			log.Error(err),
		).Error("could not sign object")
		return errors.Wrap(err, errors.New("could not sign object"))
	}

	for _, f := range fs {
		err := r.exchange.Send(ctx, o, peer.LookupByKey(f), opts...)
		if err != nil {
			logger.Debug("could not send request", log.Error(err))
		}
	}
	return nil
}

func (r *Discoverer) withoutOwnPeer(ps []*peer.Peer) []*peer.Peer {
	lp := r.local.GetPeerPublicKey().String()
	pm := map[string]*peer.Peer{}
	for _, p := range ps {
		pm[p.Signature.Signer.String()] = p
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

func (r *Discoverer) withoutOwnFingerprint(ps []crypto.PublicKey) []crypto.PublicKey {
	lp := r.local.GetPeerPublicKey().String()
	pm := map[string]crypto.PublicKey{}
	for _, p := range ps {
		pm[p.String()] = p
	}
	nps := []crypto.PublicKey{}
	for f, p := range pm {
		if f == lp {
			continue
		}
		nps = append(nps, p)
	}
	return nps
}

// func matchPeerWithLookupFilters(p *peer.Peer, fs ...peer.LookupFilter) bool {
// 	for _, f := range fs {
// 		if f(p) == false {
// 			return false
// 		}
// 	}
// 	return true
// }

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
