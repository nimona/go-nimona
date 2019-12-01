package hyperspace

import (
	"time"

	"nimona.io/internal/rand"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/discovery/bloom"
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
		context  context.Context
		store    *Store
		exchange exchange.Exchange
		local    *peer.LocalPeer
	}
)

// NewDiscoverer returns a new hyperspace discoverer
func NewDiscoverer(
	ctx context.Context,
	exc exchange.Exchange,
	local *peer.LocalPeer,
	bootstrapAddresses []string,
) (*Discoverer, error) {
	r := &Discoverer{
		context:  ctx,
		store:    NewStore(),
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
	opts ...discovery.LookupOption,
) (
	[]*peer.Peer,
	error,
) {
	opt := discovery.ParseLookupOptions(opts...)

	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.Lookup"),
	)
	logger.Debug("looking up")

	eps := []*peer.Peer{}
	if !opt.Local {
		// nolint: errcheck
		eps, _ = r.lookup(
			ctx,
			bloom.New(opt.Lookups...),
			opt.Filters,
		)
	}

	lps := r.store.Get(bloom.New(opt.Lookups...))
	if len(lps) > 0 {
		logger.Debug(
			"found peers in store",
			log.Int("n", len(lps)),
			log.Any("peers", lps),
		)
	}

	ps := r.withoutOwnPeer(append(eps, lps...))

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
		r.handlePeer(ctx, v, e)
	case peerLookupRequestType:
		v := &peer.LookupRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerLookup(ctx, v, e)
	}
	return nil
}

func (r *Discoverer) handlePeer(
	ctx context.Context,
	p *peer.Peer,
	e *exchange.Envelope,
) {
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.handlePeer"),
		log.String("peer.fingerprint", e.Sender.String()),
		log.Strings("peer.addresses", p.Addresses),
	)
	logger.Debug("adding peer to store")
	r.store.AddPeer(p)
	logger.Debug("added peer to store")
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

	cps := r.store.GetClosest(q.Bloom)
	res := &peer.LookupResponse{
		Nonce: q.Nonce,
		Peers: []crypto.PublicKey{},
	}

	sig, err := crypto.NewSignature(r.local.GetPeerPrivateKey(), res.ToObject())
	if err != nil {
		logger.Error("could not sign lookup res", log.Error(err))
		return
	}

	res.Signature = sig

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
			e.Sender.Address(),
			exchange.WithLocalDiscoveryOnly(),
			exchange.WithAsync(),
		)
		if err != nil {
			logger.Debug("could not send peer",
				log.Error(err),
			)
			continue
		}
		res.Peers = append(res.Peers, p.PublicKey())
	}
	err = r.exchange.Send(
		ctx,
		res.ToObject(),
		e.Sender.Address(),
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
	filters []discovery.LookupFilter,
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

	// allows us to match peers with peerRequest
	peerMatchesRequest := func(p *peer.Peer) bool {
		return matchPeerWithLookupFilters(p, filters...)
	}

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
					recipient.Address(),
					exchange.WithLocalDiscoveryOnly(),
					// exchange.WithAsync(),
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
			for _, r := range res.Peers {
				recipients <- r
			}
			recipientsResponded <- res.Signature.Signer
		}
	}()

	// find and ask "closest" peers
	go func() {
		cps := r.store.GetClosest(r.local.GetSignedPeer().Bloom)
		cps = r.withoutOwnPeer(cps)
		for _, p := range cps {
			if peerMatchesRequest(p) {
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
		// exchange.WithAsync(),
	}
	q := &peer.LookupRequest{
		Nonce: rand.String(12),
		Bloom: r.local.GetSignedPeer().Bloom,
	}
	o := q.ToObject()
	for _, addr := range bootstrapAddresses {
		logger.Debug("connecting to bootstrap", log.String("address", addr))
		err := r.exchange.Send(ctx, o, addr, opts...)
		if err != nil {
			logger.Debug("could not send request to bootstrap", log.Error(err))
		}
		err = r.exchange.Send(ctx, r.local.GetSignedPeer().ToObject(), addr, opts...)
		if err != nil {
			logger.Debug("could not send own peer to bootstrap", log.Error(err))
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
	cps := r.store.GetClosest(cb.Bloom)
	fs := []crypto.PublicKey{}
	for _, c := range cps {
		fs = append(fs, c.Signature.Signer)
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
		addr := f.Address()
		err := r.exchange.Send(ctx, o, addr, opts...)
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

func matchPeerWithLookupFilters(p *peer.Peer, fs ...discovery.LookupFilter) bool {
	for _, f := range fs {
		if f(p) == false {
			return false
		}
	}
	return true
}
