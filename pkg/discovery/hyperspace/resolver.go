package hyperspace

import (
	"time"

	"go.uber.org/zap"

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

	exc.Handle("/peer.request", r.handleObject)
	exc.Handle("/peer", r.handleObject)

	r.store.Add(local.GetPeerInfo())
	r.bootstrap(ctx, bootstrapAddresses)

	return r, nil
}

// FindByFingerprint finds and returns peer infos from a fingerprint
func (r *Discoverer) FindByFingerprint(
	ctx context.Context,
	fingerprint string,
	opts ...discovery.Option,
) ([]*peer.PeerInfo, error) {
	opt := discovery.ParseOptions(opts...)

	logger := log.Logger(ctx).With(
		zap.String("method", "resolver/FindByFingerprint"),
		zap.String("peerinfo.fingerprint", fingerprint),
	)
	logger.Debug("trying to find peer by fingerprint")

	eps := r.store.FindByFingerprint(fingerprint)
	if len(eps) > 0 {
		logger.Debug(
			"found peers in store",
			zap.Int("n", len(eps)),
		)

		return eps, nil
	}

	if opt.Local {
		return nil, nil
	}

	peers, _ := r.LookupPeerInfo(ctx, &peer.PeerInfoRequest{
		Keys: []string{
			fingerprint,
		},
	})

	peers = r.store.FindByFingerprint(fingerprint)
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

	r.LookupPeerInfo(ctx, &peer.PeerInfoRequest{
		ContentIDs: []string{
			contentHash,
		},
	})

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
	logger := log.Logger(ctx).With(
		zap.String("method", "resolver/handlePeerInfo"),
		zap.String("peerinfo.fingerprint", p.Fingerprint()),
		zap.Strings("peerinfo.addresses", p.Addresses),
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
	logger := log.Logger(ctx).With(
		zap.String("method", "resolver/handlePeerInfoRequest"),
		zap.String("e.sender", e.Sender.Fingerprint()),
		zap.String("e.requestID", e.RequestID),
		zap.Strings("query.contentIDs", q.ContentIDs),
		zap.Strings("query.contentTypes", q.ContentTypes),
		zap.Strings("query.keys", q.Keys),
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
			zap.String("address", addr),
			zap.String("peer", p.Fingerprint()),
		)
		err := r.exchange.Send(ctx, p.ToObject(), addr, opts...)
		if err != nil {
			logger.Debug("handleProviderRequest could not send object",
				zap.Error(err),
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
	logger := log.Logger(ctx).With(
		zap.String("method", "resolver/LookupPeerInfo"),
		zap.Strings("query.contentIDs", q.ContentIDs),
		zap.Strings("query.contentTypes", q.ContentTypes),
		zap.Strings("query.keys", q.Keys),
	)
	o := q.ToObject()
	ps := r.store.FindClosest(q)
	out := make(chan *exchange.Envelope, 10)
	rctx := context.FromContext(ctx)
	opts := []exchange.Option{
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithResponse(context.GetCorrelationID(rctx), out),
	}
	logger.Debug("found peers to ask", zap.Int("n", len(ps)))
	for _, p := range ps {
		r.exchange.Send(ctx, o, "peer:"+p.Fingerprint(), opts...)
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
				zap.String("res.type", res.Payload.GetType()),
				zap.String("res.sender", res.Sender.Fingerprint()),
			)
			r.handleObject(res)
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
	logger := log.Logger(ctx)
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
			logger.Debug("bootstrap could not send request", zap.Error(err))
		}
		err = r.exchange.Send(ctx, r.local.GetPeerInfo().ToObject(), addr, opts...)
		if err != nil {
			logger.Debug("bootstrap could not send self", zap.Error(err))
		}
	}
	return nil
}
