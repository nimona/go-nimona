package hyperspace

import (
	"time"

	"nimona.io/internal/context"
	"nimona.io/internal/log"
	"nimona.io/pkg/discovery"
	"nimona.io/pkg/net"
	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object/exchange"
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

	if _, err := exc.Handle("/peer.request", r.handleObject); err != nil {
		return nil, err
	}
	if _, err := exc.Handle("/peer", r.handleObject); err != nil {
		return nil, err
	}

	r.store.Add(local.GetPeerInfo())
	if err := r.bootstrap(ctx, bootstrapAddresses); err != nil {
		return nil, err
	}

	return r, nil
}

// FindByFingerprint finds and returns peer infos from a fingerprint
func (r *Discoverer) FindByFingerprint(
	ctx context.Context,
	fingerprint string,
	opts ...discovery.Option,
) ([]*peer.PeerInfo, error) {
	opt := discovery.ParseOptions(opts...)

	logger := log.FromContext(ctx).With(
		log.String("method", "resolver/FindByFingerprint"),
		log.String("peerinfo.fingerprint", fingerprint),
	)
	logger.Debug("trying to find peer by fingerprint")

	eps := r.store.FindByFingerprint(fingerprint)
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
		Keys: []string{
			fingerprint,
		},
	}); err != nil {
		return nil, err
	}

	peers := r.store.FindByFingerprint(fingerprint)
	return peers, nil
}

// FindByContent finds and returns peer infos from a content hash
func (r *Discoverer) FindByContent(
	ctx context.Context,
	contentHash string,
	opts ...discovery.Option,
) ([]*peer.PeerInfo, error) {
	opt := discovery.ParseOptions(opts...)

	eps := r.store.FindByContent(contentHash)
	if len(eps) > 0 {
		return eps, nil
	}

	if opt.Local {
		return nil, nil
	}

	if _, err := r.LookupPeerInfo(ctx, &peer.PeerInfoRequest{
		ContentIDs: []string{
			contentHash,
		},
	}); err != nil {
		return nil, err
	}

	eps = r.store.FindByContent(contentHash)
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
		log.String("method", "resolver/handlePeerInfo"),
		log.String("peerinfo.fingerprint", p.Fingerprint()),
		log.Strings("peerinfo.addresses", p.Addresses),
	)
	logger.Debug("adding peerinfo to store")
	r.store.Add(p)
}

func (r *Discoverer) handlePeerInfoRequest(
	ctx context.Context,
	q *peer.PeerInfoRequest,
	e *exchange.Envelope,
) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "resolver/handlePeerInfoRequest"),
		log.String("e.sender", e.Sender.Fingerprint()),
		log.String("e.requestID", e.RequestID),
		log.Strings("query.contentIDs", q.ContentIDs),
		log.Strings("query.contentTypes", q.ContentTypes),
		log.Strings("query.keys", q.Keys),
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

	pm := map[string]*peer.PeerInfo{}
	for _, p := range cps {
		pm[p.Fingerprint()] = p
	}

	fps := []*peer.PeerInfo{}
	for _, p := range pm {
		fps = append(fps, p)
	}

	opts := []exchange.Option{
		exchange.WithLocalDiscoveryOnly(),
		exchange.AsResponse(e.RequestID),
	}

	addr := "peer:" + e.Sender.Fingerprint()
	for _, p := range fps {
		logger.Debug("responding with peer",
			log.String("address", addr),
			log.String("peer", p.Fingerprint()),
		)
		err := r.exchange.Send(ctx, p.ToObject(), addr, opts...)
		if err != nil {
			logger.Debug("handleProviderRequest could not send object",
				log.Error(err),
			)
		}
	}
}

// LookupPeerInfo does a network lookup given a query
func (r *Discoverer) LookupPeerInfo(
	ctx context.Context,
	q *peer.PeerInfoRequest,
) ([]*peer.PeerInfo, error) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "resolver/LookupPeerInfo"),
		log.Strings("query.contentIDs", q.ContentIDs),
		log.Strings("query.contentTypes", q.ContentTypes),
		log.Strings("query.keys", q.Keys),
	)
	o := q.ToObject()
	ps := r.store.FindClosest(q)
	out := make(chan *exchange.Envelope, 10)
	rctx := context.FromContext(ctx)
	opts := []exchange.Option{
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithResponse(context.GetCorrelationID(rctx), out),
	}
	logger.Debug("found peers to ask", log.Int("n", len(ps)))
	for _, p := range ps {
		err := r.exchange.Send(ctx, o, "peer:"+p.Fingerprint(), opts...)
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
				log.String("res.sender", res.Sender.Fingerprint()),
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
		Keys: []string{
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
