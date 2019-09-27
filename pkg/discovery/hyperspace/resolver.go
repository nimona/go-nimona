package hyperspace

import (
	"time"

	"nimona.io/pkg/object"

	"nimona.io/pkg/context"
	"nimona.io/internal/errors"
	"nimona.io/pkg/log"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/discovery/hyperspace/bloom"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/net"
	"nimona.io/pkg/peer"
)

//go:generate $GOBIN/genny -in=../../../internal/generator/synclist/synclist.go -out=synclist_content_hashes_generated.go -pkg hyperspace gen "KeyType=string"

type (
	// Discoverer hyperspace
	Discoverer struct {
		context  context.Context
		store    *Store
		net      net.Network
		exchange exchange.Exchange
		local    *peer.LocalPeer

		contentHashes *StringSyncList
	}
)

// NewDiscoverer returns a new hyperspace discoverer
func NewDiscoverer(
	ctx context.Context,
	network net.Network,
	exc exchange.Exchange,
	local *peer.LocalPeer,
	bootstrapAddresses []string,
) (*Discoverer, error) {
	r := &Discoverer{
		context:  ctx,
		store:    NewStore(),
		net:      network,
		local:    local,
		exchange: exc,

		contentHashes: &StringSyncList{},
	}

	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/Discoverer"),
	)

	if _, err := exc.Handle("nimona.io/discovery/hyperspace/bloom", r.handleObject); err != nil {
		return nil, err
	}

	if _, err := exc.Handle("nimona.io/discovery/hyperspace/bloom.request", r.handleObject); err != nil {
		return nil, err
	}

	if _, err := exc.Handle("nimona.io/discovery/peer", r.handleObject); err != nil {
		return nil, err
	}

	if _, err := exc.Handle("nimona.io/discovery/peer.request", r.handleObject); err != nil {
		return nil, err
	}

	r.local.OnContentHashesUpdated(func(hashes []string) {
		nctx, _ := context.WithTimeout(ctx, time.Second*5)
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

	// r.store.Add(local.GetSignedPeer())
	go func() {
		if err := r.bootstrap(ctx, bootstrapAddresses); err != nil {
			logger.Error("could not bootstrap", log.Error(err))
		}
		if err := r.publishContentHashes(ctx); err != nil {
			logger.Error("could not publish initial content hashes", log.Error(err))
		}
	}()

	return r, nil
}

// FindByFingerprint finds and returns peer infos from a fingerprint
func (r *Discoverer) FindByFingerprint(
	ctx context.Context,
	fingerprint crypto.Fingerprint,
	opts ...discovery.Option,
) ([]*peer.Peer, error) {
	opt := discovery.ParseOptions(opts...)

	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.FindByFingerprint"),
		log.Any("peer.fingerprint", fingerprint),
	)
	logger.Debug("trying to find peer by fingerprint")

	eps := r.store.FindByFingerprint(fingerprint)
	eps = r.withoutOwnPeer(eps)
	if len(eps) > 0 {
		logger.Debug(
			"found peers in store",
			log.Int("n", len(eps)),
		)
		return eps, nil
	}

	if opt.Local {
		return nil, nil
	}

	if _, err := r.LookupPeer(ctx, &peer.PeerRequest{
		Keys: []crypto.Fingerprint{
			fingerprint,
		},
	}); err != nil {
		return nil, err
	}

	eps = r.store.FindByFingerprint(fingerprint)
	eps = r.withoutOwnPeer(eps)
	return eps, nil
}

// FindByContent finds and returns peer infos from a content hash
func (r *Discoverer) FindByContent(
	ctx context.Context,
	contentHash string,
	opts ...discovery.Option,
) ([]crypto.Fingerprint, error) {
	opt := discovery.ParseOptions(opts...)

	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.FindByContent"),
		log.Any("contentHash", contentHash),
		log.Any("opts", opts),
	)
	logger.Debug("trying to find peer by contentHash")

	b := bloom.NewBloom(contentHash)
	cs := r.store.FindByContent(b)
	fs := []crypto.Fingerprint{}
	if len(cs) > 0 {
		for _, b := range cs {
			fs = append(fs, b.Signature.PublicKey.Fingerprint())
		}
		logger.With(
			log.Int("n", len(cs)),
			log.Any("fingerprints", cs),
		).Debug("found n fingerprints")
		return fs, nil
	}

	if opt.Local {
		return nil, nil
	}

	logger.Debug("looking up peers")

	fs, err := r.LookupContentProvider(ctx, b)
	if err != nil {
		logger.With(
			log.Error(err),
		).Debug("failed to look up peers")
		return nil, err
	}

	logger.With(
		log.Int("n", len(fs)),
	).Debug("found n peers")
	return r.withoutOwnFingerprint(fs), nil
}

func (r *Discoverer) handleObject(e *exchange.Envelope) error {
	// attempt to recover correlation id from request id
	ctx := r.context
	if e.RequestID != "" {
		ctx = context.New(
			context.WithCorrelationID(e.RequestID),
		)
	}

	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.handleObject"),
	)

	// handle payload
	o := e.Payload
	switch o.GetType() {
	case peer.PeerRequestType:
		v := &peer.PeerRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerRequest(ctx, v, e)
	case peer.PeerType:
		v := &peer.Peer{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeer(ctx, v)
	case ContentHashesBloomType:
		logger := logger.With(
			log.String("e.sender", e.Sender.Fingerprint().String()),
			log.String("o.type", o.GetType()),
		)
		v := &ContentHashesBloom{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		if v.Signature == nil || v.Signature.PublicKey == nil {
			logger.Warn("object has no signature")
			return nil
		}
		r.handleProvider(ctx, v)
	case ContentHashesBloomRequestType:
		v := &ContentHashesBloomRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handleProviderRequest(ctx, v, e)
	}
	return nil
}

func (r *Discoverer) handlePeer(
	ctx context.Context,
	p *peer.Peer,
) {
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.handlePeer"),
		log.String("peer.fingerprint", p.Fingerprint().String()),
		log.Strings("peer.addresses", p.Addresses),
	)
	logger.Debug("adding peer to store")
	r.store.AddPeer(p)
	logger.Debug("added peer to store")
}

func (r *Discoverer) handlePeerRequest(
	ctx context.Context,
	q *peer.PeerRequest,
	e *exchange.Envelope,
) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.handlePeerRequest"),
		log.String("e.sender", e.Sender.Fingerprint().String()),
		log.String("e.requestID", e.RequestID),
		log.Strings("query.contentTypes", q.ContentTypes),
		log.Any("query.keys", q.Keys),
	)

	logger.Debug("handling peer request")

	cps := []*peer.Peer{}
	for _, k := range q.Keys {
		cps = append(cps, r.store.FindClosestPeer(k)...)
	}

	// TODO doesn't the above already find the exact peers?
	for _, f := range q.Keys {
		ps := r.store.FindByFingerprint(f)
		cps = append(cps, ps...)
	}

	cps = r.withoutOwnPeer(cps)

	for _, p := range cps {
		logger.Debug("responding with peer",
			log.String("peer", p.Fingerprint().String()),
		)
		err := e.Respond(p.ToObject())
		if err != nil {
			logger.Debug("could not send object",
				log.Error(err),
			)
		}
	}
	logger.Debug("handling done")
}

func (r *Discoverer) handleProvider(
	ctx context.Context,
	p *ContentHashesBloom,
) {
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.handleProvider"),
		log.String("p.fingerprint", p.Signature.PublicKey.Fingerprint().String()),
	)
	logger.Debug("adding provider to store")
	r.store.AddContentHashes(p)
	logger.Debug("added provider to store")
}

func (r *Discoverer) handleProviderRequest(
	ctx context.Context,
	q *ContentHashesBloomRequest,
	e *exchange.Envelope,
) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.handleProviderRequest"),
		log.String("e.sender", e.Sender.Fingerprint().String()),
		log.String("e.requestID", e.RequestID),
		log.Any("query.bloom", q.BloomFilter),
	)

	logger.Debug("handling content request")

	cps := r.store.FindClosestContentProvider(q)

	oc, err := r.getOwnContentHashesBloom()
	if err == nil && oc != nil {
		cps = append(cps, oc)
	}

	for _, p := range cps {
		pf := ""
		if p.Signature != nil {
			pf = p.Signature.PublicKey.Fingerprint().String()
		}
		logger.Debug("responding with content hash bloom",
			log.Any("bloom", p.Bloom),
			log.String("fingerprint", pf),
		)
		err := e.Respond(p.ToObject())
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
	q bloom.Bloomer,
) ([]crypto.Fingerprint, error) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.LookupContentProvider"),
		log.Any("query.bloom", q),
	)
	cq := &ContentHashesBloomRequest{
		BloomFilter: q.Bloom(),
	}
	o := cq.ToObject()
	cs := r.store.FindClosestContentProvider(q)
	fs := []crypto.Fingerprint{}
	for _, c := range cs {
		fs = append(fs, c.Signature.PublicKey.Fingerprint())
	}
	ps := r.store.FindClosestPeer(r.local.GetFingerprint())
	for _, p := range ps {
		fs = append(fs, p.Fingerprint())
	}
	if len(fs) == 0 {
		logger.Debug("couldn't find peers to ask")
		return nil, errors.New("no peers to ask")
	}

	out := make(chan *exchange.Envelope, 10)
	rctx, _ := context.WithTimeout(ctx, time.Second*5)
	opts := []exchange.Option{
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithResponse(context.GetCorrelationID(rctx), out),
	}
	logger.Debug("found peers to ask", log.Int("n", len(fs)))
	for _, f := range fs {
		go func(f crypto.Fingerprint) {
			logger.With(
				log.String("peer", f.String()),
			).Debug("asking peer")
			err := r.exchange.Send(
				ctx,
				object.Copy(o),
				f.Address(),
				opts...,
			)
			if err != nil {
				logger.Debug("could not lookup peer", log.Error(err))
				return
			}
			logger.With(
				log.String("peer", f.String()),
			).Debug("asked peer")
		}(f)
	}
	// TODO(geoah) better timeout
	t := time.NewTimer(time.Second * 5)
loop:
	for {
		select {
		case <-t.C:
			logger.Debug("timer done, giving up")
			break loop

		case <-ctx.Done():
			break loop

		case res := <-out:
			logger.Debug("got loopkup response",
				log.String("res.type", res.Payload.GetType()),
				log.String("res.sender", res.Sender.Fingerprint().String()),
			)
			if err := r.handleObject(res); err != nil {
				logger.Debug("could not handle object", log.Error(err))
			}
			// if res.Payload.GetType() == ContentHashesBloomType {
			// 	v := &ContentHashesBloom{}
			// 	if err := v.FromObject(res.Payload); err == nil {
			// 		fingerprints = append(
			// 			fingerprints,
			// 			v.Signature.PublicKey.Fingerprint(),
			// 		)
			// 	}
			// 	break
			// }
		}
	}
	cs = r.store.FindByContent(q)
	fs = []crypto.Fingerprint{}
	for _, c := range cs {
		fs = append(fs, c.Signature.PublicKey.Fingerprint())
	}
	logger.With(
		log.Int("n", len(fs)),
	).Debug("done, found n providers")
	return fs, nil
}

// LookupPeer does a network lookup given a query
func (r *Discoverer) LookupPeer(
	ctx context.Context,
	q *peer.PeerRequest,
) ([]*peer.Peer, error) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.LookupPeer"),
		log.Strings("query.contentTypes", q.ContentTypes),
		log.Any("query.keys", q.Keys),
	)
	o := q.ToObject()
	// TODO support multiple keys
	ps := r.store.FindClosestPeer(q.Keys[0])
	out := make(chan *exchange.Envelope, 10)
	rctx := context.FromContext(ctx)
	opts := []exchange.Option{
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithResponse(context.GetCorrelationID(rctx), out),
	}
	ps = r.withoutOwnPeer(ps)
	if len(ps) == 0 {
		logger.Debug("couldn't find peers to ask")
		return nil, errors.New("no peers to ask")
	}

	logger.Debug("found peers to ask", log.Int("n", len(ps)))
	for _, p := range ps {
		err := r.exchange.Send(ctx, o, "peer:"+p.Fingerprint().String(), opts...)
		if err != nil {
			logger.Debug("could not lookup peer", log.Error(err))
		}
	}
	peers := []*peer.Peer{}
	// TODO(geoah) better timeout
	t := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-t.C:
			return peers, nil

		case <-ctx.Done():
			return peers, nil

		case res := <-out:
			logger.Debug("got loopkup response",
				log.String("res.type", res.Payload.GetType()),
				log.String("res.sender", res.Sender.Fingerprint().String()),
			)
			if err := r.handleObject(res); err != nil {
				logger.Debug("could not handle object", log.Error(err))
			}
			if res.Payload.GetType() == peer.PeerType {
				v := &peer.Peer{}
				if err := v.FromObject(res.Payload); err == nil {
					peers = append(peers, v)
				}
				return peers, nil
			}
		}
	}
	return peers, nil
}

func (r *Discoverer) bootstrap(
	ctx context.Context,
	bootstrapAddresses []string,
) error {
	logger := log.FromContext(ctx)
	key := r.local.GetPeerKey()
	opts := []exchange.Option{
		exchange.WithLocalDiscoveryOnly(),
	}
	q := &peer.PeerRequest{
		Keys: []crypto.Fingerprint{
			key.Fingerprint(),
		},
	}
	o := q.ToObject()
	for _, addr := range bootstrapAddresses {
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

func (r *Discoverer) getContentHashes() []string {
	cIDs := []string{}
	r.contentHashes.Range(func(k string) bool {
		cIDs = append(cIDs, k)
		return true
	})
	return cIDs
}

func (r *Discoverer) getOwnContentHashesBloom() (*ContentHashesBloom, error) {
	cs := r.getContentHashes()
	b := bloom.NewBloom(cs...)
	cb := &ContentHashesBloom{
		BloomFilter: b,
	}

	o := cb.ToObject()
	if err := crypto.Sign(o, r.local.GetPeerKey()); err != nil {
		return nil, errors.Wrap(err, errors.New("could not sign object"))
	}

	err := cb.FromObject(o)
	return cb, err
}

func (r *Discoverer) publishContentHashes(
	ctx context.Context,
) error {
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/Discoverer.publishContentHashes"),
	)
	cs := r.getContentHashes()
	b := bloom.NewBloom(cs...)

	cps := r.store.FindClosestContentProvider(b)
	fs := []crypto.Fingerprint{}
	for _, c := range cps {
		fs = append(fs, c.Signature.PublicKey.Fingerprint())
	}
	ps := r.store.FindClosestPeer(r.local.GetFingerprint())
	for _, p := range ps {
		fs = append(fs, p.Fingerprint())
	}
	if len(fs) == 0 {
		logger.Debug("couldn't find peers to tell")
		return errors.New("no peers to tell")
	}

	logger.With(
		log.Int("n", len(fs)),
		log.Any("bloom", b),
	).Debug("trying to tell n peers")

	cb := &ContentHashesBloom{
		BloomFilter: b,
	}
	opts := []exchange.Option{
		exchange.WithLocalDiscoveryOnly(),
	}

	o := cb.ToObject()
	if err := crypto.Sign(o, r.local.GetPeerKey()); err != nil {
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
	lp := r.local.GetFingerprint().String()
	pm := map[crypto.Fingerprint]*peer.Peer{}
	for _, p := range ps {
		pm[p.Fingerprint()] = p
	}
	nps := []*peer.Peer{}
	for f, p := range pm {
		if f.String() == lp {
			continue
		}
		nps = append(nps, p)
	}
	return nps
}

func (r *Discoverer) withoutOwnFingerprint(ps []crypto.Fingerprint) []crypto.Fingerprint {
	lp := r.local.GetFingerprint().String()
	pm := map[crypto.Fingerprint]crypto.Fingerprint{}
	for _, p := range ps {
		pm[p] = p
	}
	nps := []crypto.Fingerprint{}
	for f, p := range pm {
		if f.String() == lp {
			continue
		}
		nps = append(nps, p)
	}
	return nps
}
