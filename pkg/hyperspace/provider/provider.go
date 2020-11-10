package provider

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"nimona.io/pkg/context"
	"nimona.io/pkg/hyperspace"
	"nimona.io/pkg/hyperspace/peerstore"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

const (
	peerCacheTTL = time.Minute * 15
	peerCacheGC  = time.Minute * 1
)

var (
	hyperspaceAnnouncementType  = new(hyperspace.Announcement).Type()
	hyperspaceLookupRequestType = new(hyperspace.LookupRequest).Type()
)

var (
	promIncRequestsCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_hyperspace_provider_lookup_requests",
			Help: "Total number of incoming lookup requests",
		},
	)
)

type (
	Provider struct {
		context   context.Context
		network   network.Network
		peerCache *peerstore.PeerCache
		local     localpeer.LocalPeer
	}
)

func New(
	ctx context.Context,
	net network.Network,
) (*Provider, error) {
	c := peerstore.NewPeerCache(peerCacheGC, "nimona_hyperspace_provider")
	p := &Provider{
		context:   ctx,
		network:   net,
		local:     net.LocalPeer(),
		peerCache: c,
	}

	go network.HandleEnvelopeSubscription(
		p.network.Subscribe(),
		p.handleObject,
	)

	return p, nil
}

func (p *Provider) Put(
	prs ...*hyperspace.Announcement,
) {
	for _, pr := range prs {
		p.peerCache.Put(pr, peerCacheTTL)
	}
}

func (p *Provider) handleObject(
	e *network.Envelope,
) error {
	ctx := p.context

	o := e.Payload
	switch o.Type {
	case hyperspaceAnnouncementType:
		v := &hyperspace.Announcement{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		p.handleAnnouncement(ctx, v)
	case hyperspaceLookupRequestType:
		v := &hyperspace.LookupRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		p.handlePeerLookup(ctx, v, e)
		return nil
	}

	return nil
}

func (p *Provider) handleAnnouncement(
	ctx context.Context,
	ann *hyperspace.Announcement,
) {
	logger := log.FromContext(ctx).With(
		log.String("method", "provider.handleAnnouncement"),
		log.String("peer.publicKey", ann.Peer.PublicKey.String()),
		log.Strings("peer.addresses", ann.Peer.Addresses),
	)
	// TODO check if we've already received this peer, and if not forward it
	// to the other hyperspace providers
	logger.Debug("adding peer to cache")
	p.peerCache.Put(ann, peerCacheTTL)
}

func (p *Provider) handlePeerLookup(
	ctx context.Context,
	q *hyperspace.LookupRequest,
	e *network.Envelope,
) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "provider.handlePeerLookup"),
		log.String("e.sender", e.Sender.String()),
		log.Any("q.vector", q.QueryVector),
		log.Any("o.signer", e.Payload.Metadata.Signature.Signer),
	)

	promIncRequestsCounter.Inc()

	logger.Debug("handling peer lookup")

	ans := p.peerCache.Lookup(hyperspace.Bloom(q.QueryVector))

	ctx = context.New(
		context.WithParent(ctx),
	)

	res := &hyperspace.LookupResponse{
		Metadata: object.Metadata{
			Owner: p.local.GetPrimaryPeerKey().PublicKey(),
		},
		Nonce:         q.Nonce,
		QueryVector:   q.QueryVector,
		Announcements: ans,
	}

	pr := &peer.ConnectionInfo{
		PublicKey: e.Sender,
	}

	err := p.network.Send(
		ctx,
		res.ToObject(),
		pr,
	)
	if err != nil {
		logger.Debug("could not send lookup response",
			log.Error(err),
		)
		return
	}

	logger.With(
		log.Int("n", len(ans)),
	).Debug("handling done, sent n peers")
}
