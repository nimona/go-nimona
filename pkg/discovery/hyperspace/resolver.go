package hyperspace

import (
	"sort"
	"sync"

	"nimona.io/pkg/object"

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
) (<-chan *peer.Peer, error) {
	opt := peer.ParseLookupOptions(opts...)

	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.Lookup"),
	)
	logger.Debug("looking up")

	bloom := bloom.New(opt.Lookups...)

	// create channel to keep peers we find
	peers := make(chan *peer.Peer, 100)

	// send content requests to recipients
	req := &peer.LookupRequest{
		Nonce: rand.String(12),
		Bloom: bloom,
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
		for {
			e, err := resSub.Next()
			if err != nil {
				break
			}
			peerLookupResponses <- e
		}
	}()

	// create channel for the peers we need to ask
	recipients := make(chan crypto.PublicKey)
	// keep a record of who we asked and who has responded
	recipientsResponded := &sync.Map{}

	go func() {
		for {
			recipient := <-recipients
			// check if we've already asked them
			if _, asked := recipientsResponded.Load(recipient); asked {
				continue
			}
			// else mark them as already been asked, but not responded
			recipientsResponded.Store(recipient, false)
			// and finally ask them
			err := r.exchange.Send(
				ctx,
				reqObject,
				peer.LookupByOwner(recipient),
				exchange.WithLocalDiscoveryOnly(),
				exchange.WithAsync(),
			)
			if err != nil {
				logger.Debug("could send request to peer", log.Error(err))
			}
			logger.Debug("asked peer", log.String("peer", recipient.String()))
		}
	}()

	go func() {
	loop:
		for {
			select {
			case <-ctx.Done():
				logger.Debug("ctx done, giving up")
				break loop
			case e := <-peerLookupResponses:
				res := &peer.LookupResponse{}
				if err := res.FromObject(e.Payload); err != nil {
					continue
				}
				// recipientsResponded[e.Sender.String()] = true
				recipientsResponded.Store(e.Sender, true)
				for _, p := range res.Peers {
					if opt.Match(p) {
						peers <- p
					}
					recipients <- p.PublicKey()
				}
				allDone := true
				recipientsResponded.Range(func(peer, answered interface{}) bool {
					if v, ok := answered.(bool); !ok || !v {
						allDone = false
						return false
					}
					return true
				})
				if allDone {
					break loop
				}
			}
		}
		close(peers)
		resSub.Cancel()
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
	cps := getClosest(pps, bloom)
	cps = r.withoutOwnPeer(cps)
	for _, p := range cps {
		recipients <- p.PublicKey()
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
	pps := []*peer.Peer{}
	for p := range aps {
		pps = append(pps, p)
	}
	cps := getClosest(pps, q.Bloom)
	cps = append(cps, r.local.GetSignedPeer())
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
	pps := []*peer.Peer{}
	for p := range aps {
		pps = append(pps, p)
	}
	cps := getClosest(pps, cb.Bloom)
	fs := []crypto.PublicKey{}
	for _, c := range cps {
		if !c.Signature.IsEmpty() {
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
	sig, err := object.NewSignature(r.local.GetPeerPrivateKey(), o)
	if err != nil {
		logger.With(
			log.Error(err),
		).Error("could not sign object")
		return errors.Wrap(err, errors.New("could not sign object"))
	}

	o = o.SetSignature(sig)
	for _, f := range fs {
		err := r.exchange.Send(ctx, o, peer.LookupByOwner(f), opts...)
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
