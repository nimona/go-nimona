package hyperspace

import (
	"time"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/discovery/hyperspace/bloom"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

//go:generate $GOBIN/genny -in=$GENERATORS/synclist/synclist.go -out=synclist_content_hashes_generated.go -pkg hyperspace gen "KeyType=object.Hash"

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

		contentHashes *ObjectHashSyncList
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

		contentHashes: &ObjectHashSyncList{},
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

	// listen on content hash updates
	r.local.OnContentHashesUpdated(func(hashes []object.Hash) {
		nctx := context.New(context.WithTimeout(time.Second * 3))
		for _, hash := range hashes {
			r.contentHashes.Put(hash)
		}
		err := r.publishContentHashes(nctx)
		if err != nil {
			logger.Debug("could not publish content hashes",
				log.Error(err),
			)
		}
	})

	// pre-fill content hashes from local info
	for _, ch := range r.local.GetContentHashes() {
		r.contentHashes.Put(ch)
	}

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

// FindByPublicKey finds and returns peer infos from a fingerprint
func (r *Discoverer) FindByPublicKey(
	ctx context.Context,
	publicKey crypto.PublicKey,
	opts ...discovery.Option,
) ([]*peer.Peer, error) {
	opt := discovery.ParseOptions(opts...)

	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.FindByPublicKey"),
		log.Any("peer.publicKey", publicKey),
	)
	logger.Debug("trying to find peer by publicKey")

	eps := r.store.FindByPublicKey(publicKey)
	if len(eps) > 0 {
		logger.Debug(
			"found peers in store",
			log.Int("n", len(eps)),
			log.Any("peers", eps),
		)
	}

	if opt.Local {
		return eps, nil
	}

	meps, _ := r.LookupPeer(ctx, publicKey) // nolint: errcheck

	eps = append(eps, meps...)
	eps = r.withoutOwnPeer(eps)

	return eps, nil
}

// FindByContent finds and returns peer infos from a content hash
func (r *Discoverer) FindByContent(
	ctx context.Context,
	contentHash object.Hash,
	opts ...discovery.Option,
) ([]crypto.PublicKey, error) {
	opt := discovery.ParseOptions(opts...)

	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.FindByContent"),
		log.Any("contentHash", contentHash),
		log.Any("opts", opts),
	)
	logger.Debug("trying to find peer by contentHash")

	b := bloom.New(contentHash.String())
	cs := r.store.FindByContent(b)
	fs := []crypto.PublicKey{}

	for _, b := range cs {
		fs = append(fs, b.Signature.Signer)
	}
	logger.With(
		log.Any("fingerprints", cs),
	).Debug("found fingerprints in store")

	if opt.Local {
		return fs, nil
	}

	logger.Debug("looking up providers")

	mfs, err := r.LookupContentProvider(ctx, b)
	if err != nil {
		logger.With(
			log.Error(err),
		).Debug("failed to look up peers")
		return nil, err
	}

	fs = append(fs, mfs...)
	logger.With(
		log.Int("n", len(fs)),
	).Debug("found n peers")
	return r.withoutOwnFingerprint(fs), nil
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
		log.Any("query.key", q.Key),
	)

	logger.Debug("handling peer request")

	cps := r.store.FindClosestPeer(q.Key)
	// TODO doesn't the above already find the exact peers?
	cps = append(cps, r.store.FindByPublicKey(q.Key)...)
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

	cps := r.store.FindClosestContentProvider(q.Bloom)
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

// LookupContentProvider searches the network for content given a bloom filter
func (r *Discoverer) LookupContentProvider(
	ctx context.Context,
	q bloom.Bloom,
) ([]crypto.PublicKey, error) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.LookupContentProvider"),
		log.Any("query.bloom", q),
	)

	// create content request
	contentRequest := &peer.Lookup{
		Bloom: q.Bloom(),
	}

	// create channel to keep providers we find
	providers := make(chan crypto.PublicKey)

	// create channel for the peers we need to ask
	recipients := make(chan crypto.PublicKey)

	// send content requests to recipients
	contentRequestObject := contentRequest.ToObject()
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
				contentRequestObject,
				recipient.Address(),
				exchange.WithLocalDiscoveryOnly(),
				exchange.WithAsync(),
			)
			if err != nil {
				logger.Debug("could not lookup peer", log.Error(err))
			}
			logger.Debug("asked peer", log.String("peer", recipient.String()))
		}
	}()

	// listen for new content providers and ask them
	contentProviderHandler := func(e *exchange.Envelope) error {
		cp := &peer.Peer{}
		cp.FromObject(e.Payload) // nolint: errcheck
		if cp.Signature == nil || cp.Signature.Signer == "" {
			return nil
		}
		if intersectionCount(cp.Bloom, q) == len(q.Bloom()) {
			providers <- cp.Signature.Signer
			return nil
		}
		recipients <- cp.Signature.Signer
		return nil
	}
	providerSub := r.exchange.Subscribe(
		exchange.FilterByObjectType(peerType),
	)
	defer providerSub.Cancel()
	go exchange.HandleEnvelopeSubscription(
		providerSub,
		contentProviderHandler,
	)

	// listen for new peers and ask them
	peerHandler := func(e *exchange.Envelope) error {
		cp := &peer.Peer{}
		cp.FromObject(e.Payload) // nolint: errcheck
		if cp.Signature != nil && cp.Signature.Signer != "" {
			recipients <- cp.Signature.Signer
		}
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

	// find and ask "closest" content providers
	announcements := r.store.FindClosestContentProvider(q)
	for _, announcement := range announcements {
		recipients <- announcement.Signature.Signer
	}

	// find and ask "closest" peers
	closestPeers := r.store.FindClosestPeer(r.local.GetPeerPublicKey())
	for _, peer := range closestPeers {
		recipients <- peer.PublicKey()
	}

	// error early if there are no peers to contact
	if len(announcements) == 0 && len(closestPeers) == 0 {
		return nil, ErrNoPeersToAsk
	}

	// gather all providers until something happens
	providersList := []crypto.PublicKey{}
	t := time.NewTimer(time.Second * 5)
loop:
	for {
		select {
		case provider := <-providers:
			providersList = append(providersList, provider)
			break loop
		case <-t.C:
			logger.Debug("timer done, giving up")
			break loop
		case <-ctx.Done():
			logger.Debug("ctx done, giving up")
			break loop
		}
	}

	logger.Debug("done, found n providers", log.Int("n", len(providersList)))
	return providersList, nil
}

// LookupPeer does a network lookup given a query
func (r *Discoverer) LookupPeer(
	ctx context.Context,
	requestedPeer crypto.PublicKey,
) ([]*peer.Peer, error) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.LookupPeer"),
		log.Any("requestedPeer", requestedPeer),
	)
	logger.Debug("looking up peer")

	// create channel to keep peers we find
	peers := make(chan *peer.Peer)

	// create channel for the peers we need to ask
	recipients := make(chan crypto.PublicKey)

	// allows us to match peers with peerRequest
	peerMatchesRequest := func(peer *peer.Peer) bool {
		for _, cert := range peer.Certificates {
			if cert.Signature.Signer.Equals(requestedPeer) {
				return true
			}
		}
		return false
	}

	// send content requests to recipients
	peerRequest := &peer.Request{
		Key: requestedPeer,
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
			go func(peerRequestObject object.Object, address string) {
				err := r.exchange.Send(
					ctx,
					peerRequestObject,
					address,
					exchange.WithLocalDiscoveryOnly(),
				)
				if err != nil {
					logger.Debug("could not lookup peer", log.Error(err))
				}
			}(peerRequestObject, recipient.Address())
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
		cps := r.store.FindClosestPeer(r.local.GetPeerPublicKey())
		cps = r.withoutOwnPeer(cps)
		for _, peer := range cps {
			if peerMatchesRequest(peer) {
				peers <- peer
				continue
			}
			recipients <- peer.PublicKey()
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
	publicKey := r.local.GetPeerPublicKey()
	opts := []exchange.Option{
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithAsync(),
	}
	q := &peer.Request{
		Key: publicKey,
	}
	o := q.ToObject()
	for _, addr := range bootstrapAddresses {
		logger.Debug("connecting to bootstrap", log.String("address", addr))
		err := r.exchange.Send(ctx, o, addr, opts...)
		if err != nil {
			logger.Debug("bootstrap could not send request", log.Error(err))
		}
		err = r.exchange.Send(ctx, r.local.GetSignedPeer().ToObject(), addr, opts...)
		if err != nil {
			logger.Debug("bootstrap could not send self", log.Error(err))
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
	cps := r.store.FindClosestContentProvider(cb.Bloom)
	fs := []crypto.PublicKey{}
	for _, c := range cps {
		fs = append(fs, c.Signature.Signer)
	}
	ps := r.store.FindClosestPeer(r.local.GetPeerPublicKey())
	for _, p := range ps {
		fs = append(fs, p.Signature.Signer)
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
