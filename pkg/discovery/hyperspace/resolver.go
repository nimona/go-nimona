package hyperspace

import (
	"errors"
	"time"

	"nimona.io/internal/context"
	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/net"
	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/exchange"
)

// Discoverer hyperspace
type Discoverer struct {
	context  context.Context
	store    *Store
	net      net.Network
	exchange exchange.Exchange
	local    *net.LocalInfo
}

// NewDiscoverer returns a new hyperspace discoverer
func NewDiscoverer(
	ctx context.Context,
	network net.Network,
	exc exchange.Exchange,
	local *net.LocalInfo,
	bootstrapAddresses []string,
) (*Discoverer, error) {
	r := &Discoverer{
		context:  ctx,
		store:    NewStore(),
		net:      network,
		local:    local,
		exchange: exc,
	}

	logger := log.FromContext(ctx)

	if _, err := exc.Handle("/peer.request", r.handleObject); err != nil {
		return nil, err
	}
	if _, err := exc.Handle("/peer", r.handleObject); err != nil {
		return nil, err
	}

	// r.store.Add(local.GetPeerInfo())
	go func() {
		err := r.bootstrap(ctx, bootstrapAddresses)
		if err != nil {
			logger.Error("could not bootstrap", log.Error(err))
		}
	}()

	return r, nil
}

// FindByFingerprint finds and returns peer infos from a fingerprint
func (r *Discoverer) FindByFingerprint(
	ctx context.Context,
	fingerprint crypto.Fingerprint,
	opts ...discovery.Option,
) ([]*peer.PeerInfo, error) {
	opt := discovery.ParseOptions(opts...)

	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.FindByFingerprint"),
		log.Any("peerinfo.fingerprint", fingerprint),
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

	if _, err := r.LookupPeerInfo(ctx, &peer.PeerInfoRequest{
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
) ([]*peer.PeerInfo, error) {
	opt := discovery.ParseOptions(opts...)

	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.FindByContent"),
		log.Any("contentHash", contentHash),
		log.Any("opts", opts),
	)
	logger.Debug("trying to find peer by contentHash")

	eps := r.store.FindByContent(contentHash)
	eps = r.withoutOwnPeer(eps)
	if len(eps) > 0 {
		ps := []string{}
		for _, p := range eps {
			ps = append(ps, p.Fingerprint().String())
		}
		logger.With(
			log.Int("n", len(eps)),
			log.Strings("peers", ps),
		).Debug("found n peers")
		return eps, nil
	}

	if opt.Local {
		return nil, nil
	}

	logger.Debug("looking up peers")

	if _, err := r.LookupPeerInfo(ctx, &peer.PeerInfoRequest{
		ContentIDs: []string{
			contentHash,
		},
	}); err != nil {
		logger.With(
			log.Error(err),
		).Debug("failed to look up peers")
		return nil, err
	}

	eps = r.store.FindByContent(contentHash)
	eps = r.withoutOwnPeer(eps)
	logger.With(
		log.Int("n", len(eps)),
	).Debug("found n peers")
	return eps, nil
}

func (r *Discoverer) handleObject(e *exchange.Envelope) error {
	// attempt to recover correlation id from request id
	ctx := r.context
	if e.RequestID != "" {
		ctx = context.New(
			context.WithCorrelationID(e.RequestID),
		)
	}

	// handle payload
	o := e.Payload
	switch o.GetType() {
	case peer.PeerInfoRequestType:
		v := &peer.PeerInfoRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerInfoRequest(ctx, v, e)
	case peer.PeerInfoType:
		v := &peer.PeerInfo{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerInfo(ctx, v)
	}
	return nil
}

func (r *Discoverer) handlePeerInfo(
	ctx context.Context,
	p *peer.PeerInfo,
) {
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.handlePeerInfo"),
		log.String("peerinfo.fingerprint", p.Fingerprint().String()),
		log.Strings("peerinfo.addresses", p.Addresses),
	)
	logger.Debug("adding peerinfo to store")
	r.store.Add(p)
	logger.Debug("added peerinfo to store")
}

func (r *Discoverer) handlePeerInfoRequest(
	ctx context.Context,
	q *peer.PeerInfoRequest,
	e *exchange.Envelope,
) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.handlePeerInfoRequest"),
		log.String("e.sender", e.Sender.Fingerprint().String()),
		log.String("e.requestID", e.RequestID),
		log.Strings("query.contentIDs", q.ContentIDs),
		log.Strings("query.contentTypes", q.ContentTypes),
		log.Any("query.keys", q.Keys),
	)

	logger.Debug("handling peer info request")

	cps := r.store.FindClosest(q)

	for _, f := range q.Keys {
		ps := r.store.FindByFingerprint(f)
		cps = append(cps, ps...)
	}

	for _, c := range q.ContentIDs {
		ps := r.store.FindByContent(c)
		cps = append(cps, ps...)
	}

	cps = r.withoutOwnPeer(cps)

	for _, p := range cps {
		logger.Debug("responding with peer",
			log.String("peer", p.Fingerprint().String()),
		)
		err := e.Respond(p.ToObject())
		if err != nil {
			logger.Debug("handleProviderRequest could not send object",
				log.Error(err),
			)
		}
	}
	logger.Debug("handling done")
}

// LookupPeerInfo does a network lookup given a query
func (r *Discoverer) LookupPeerInfo(
	ctx context.Context,
	q *peer.PeerInfoRequest,
) ([]*peer.PeerInfo, error) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "hyperspace/resolver.LookupPeerInfo"),
		log.Strings("query.contentIDs", q.ContentIDs),
		log.Strings("query.contentTypes", q.ContentTypes),
		log.Any("query.keys", q.Keys),
	)
	o := q.ToObject()
	ps := r.store.FindClosest(q)
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
	peers := []*peer.PeerInfo{}
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
			if res.Payload.GetType() == peer.PeerInfoType {
				v := &peer.PeerInfo{}
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
	q := &peer.PeerInfoRequest{
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
		err = r.exchange.Send(ctx, r.local.GetPeerInfo().ToObject(), addr, opts...)
		if err != nil {
			logger.Debug("bootstrap could not send self", log.Error(err))
		}
	}
	return nil
}

func (r *Discoverer) withoutOwnPeer(ps []*peer.PeerInfo) []*peer.PeerInfo {
	lp := r.local.GetFingerprint().String()
	pm := map[crypto.Fingerprint]*peer.PeerInfo{}
	for _, p := range ps {
		pm[p.Fingerprint()] = p
	}
	nps := []*peer.PeerInfo{}
	for f, p := range pm {
		if f.String() == lp {
			continue
		}
		nps = append(nps, p)
	}
	return nps
}
