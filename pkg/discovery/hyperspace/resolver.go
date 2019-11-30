package hyperspace

import (
	"time"

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
	peerType        = new(peer.Peer).GetType()
	peerLookupType  = new(peer.Lookup).GetType()
	peerRequestType = new(peer.Request).GetType()
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
			peerRequestType,
			peerLookupType,
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
	case peerRequestType:
		v := &peer.Request{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerRequest(ctx, v, e)
	case peerType:
		v := &peer.Peer{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeer(ctx, v, e)
	case peerLookupType:
		v := &peer.Lookup{}
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

func (r *Discoverer) handlePeerRequest(
	ctx context.Context,
	q *peer.Request,
	e *exchange.Envelope,
) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.handlePeerRequest"),
		log.String("e.sender", e.Sender.String()),
		log.Any("query.bloom", q.Bloom),
	)

	logger.Debug("handling peer request")

	cps := r.store.GetClosest(q.Bloom)
	cps = r.withoutOwnPeer(cps)

	for _, p := range cps {
		logger.Debug("responding with peer",
			log.String("peer", p.Signature.Signer.String()),
		)
		err := r.exchange.Send(
			context.New(
				context.WithParent(ctx),
			),
			p.ToObject(),
			e.Sender.Address(),
			exchange.WithLocalDiscoveryOnly(),
		)
		if err != nil {
			logger.Debug("could not send object",
				log.Error(err),
			)
		}
	}
	logger.Debug("handling done")
}

func (r *Discoverer) handlePeerLookup(
	ctx context.Context,
	q *peer.Lookup,
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
	cps = append(cps, r.local.GetSignedPeer())

	for _, p := range cps {
		pf := ""
		if p.Signature != nil {
			pf = p.Signature.Signer.String()
		}
		logger.Debug("responding with content hash bloom",
			log.Any("bloom", p.Bloom),
			log.String("fingerprint", pf),
		)
		err := r.exchange.Send(
			context.New(
				context.WithParent(ctx),
			),
			p.ToObject(),
			e.Sender.Address(),
			exchange.WithLocalDiscoveryOnly(),
		)
		if err != nil {
			logger.Debug("could not send object",
				log.Error(err),
			)
		}
	}
	logger.With(
		log.Int("n", len(cps)),
	).Debug("handling done, sent n blooms")
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

	// create channel to keep peers we find
	peers := make(chan *peer.Peer)

	// create channel for the peers we need to ask
	recipients := make(chan crypto.PublicKey)

	// allows us to match peers with peerRequest
	peerMatchesRequest := func(p *peer.Peer) bool {
		return matchPeerWithLookupFilters(p, filters...)
	}

	// send content requests to recipients
	peerRequest := &peer.Request{
		Bloom: bloom,
	}
	peerRequestObject := peerRequest.ToObject()
	go func() {
		// keep a record of the peers we already asked
		recipientsAsked := map[string]bool{}
		// go through the peers we need to ask
		for recipient := range recipients {
			// check if we've already asked them
			if _, alreadyAsked := recipientsAsked[recipient.String()]; alreadyAsked {
				continue
			}
			// else mark them as already been asked
			recipientsAsked[recipient.String()] = true
			// and finally ask them
			err := r.exchange.Send(
				ctx,
				peerRequestObject,
				recipient.Address(),
				exchange.WithLocalDiscoveryOnly(),
				exchange.WithAsync(),
			)
			if err != nil {
				logger.Debug("could send request to peer", log.Error(err))
			}
			logger.Debug("asked peer", log.String("peer", recipient.String()))
		}
	}()

	// listen for new peers and ask them
	peerHandler := func(e *exchange.Envelope) error {
		cp := &peer.Peer{}
		cp.FromObject(e.Payload) // nolint: errcheck
		if cp.Signature == nil || cp.Signature.Signer == "" {
			return nil
		}
		if peerMatchesRequest(cp) {
			peers <- cp
			return nil
		}
		recipients <- cp.Signature.Signer
		return nil
	}
	peerSub := r.exchange.Subscribe(
		exchange.FilterByObjectType(peerType),
	)
	defer peerSub.Cancel()
	go exchange.HandleEnvelopeSubscription(
		peerSub,
		peerHandler,
	)

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
	q := &peer.Request{
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
