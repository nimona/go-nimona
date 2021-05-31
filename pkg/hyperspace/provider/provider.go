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
	peerCacheTTL     = time.Minute * 15
	peerCacheGC      = time.Minute * 1
	providerCacheTTL = time.Minute * 5
	providerCacheGC  = time.Minute * 1
)

var (
	hyperspaceAnnouncementType  = new(hyperspace.Announcement).Type()
	hyperspaceLookupRequestType = new(hyperspace.LookupRequest).Type()
)

var (
	announceVersion        = time.Now().Unix()
	promIncRequestsCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "nimona_hyperspace_provider_lookup_requests",
			Help: "Total number of incoming lookup requests",
		},
	)
	promIncResponsesHistogram = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name: "nimona_hyperspace_provider_lookup_response_peers",
			Help: "Number of peers in response",
		},
	)
)

type (
	Provider struct {
		context       context.Context
		network       network.Network
		peerCache     *peerstore.PeerCache
		providerCache *peerstore.PeerCache
		local         localpeer.LocalPeer
	}
)

func New(
	ctx context.Context,
	net network.Network,
	bootstrapProviders []*peer.ConnectionInfo,
) (*Provider, error) {
	p := &Provider{
		context: ctx,
		network: net,
		local:   net.LocalPeer(),
		peerCache: peerstore.NewPeerCache(
			peerCacheGC,
			"nimona_hyperspace_provider_peers",
		),
		providerCache: peerstore.NewPeerCache(
			providerCacheGC,
			"nimona_hyperspace_provider_bootstraps",
		),
	}

	go network.HandleEnvelopeSubscription(
		p.network.Subscribe(),
		func(e *network.Envelope) error {
			go p.handleObject(e)
			return nil
		},
	)

	for _, ci := range bootstrapProviders {
		p.providerCache.Put(
			&hyperspace.Announcement{
				ConnectionInfo: ci,
				PeerCapabilities: []string{
					hyperspaceAnnouncementType,
					hyperspaceLookupRequestType,
				},
			},
			providerCacheTTL,
		)
	}

	go func() {
		p.announceSelf()
		announceTimer := time.NewTicker(30 * time.Second)
		for range announceTimer.C {
			p.announceSelf()
		}
	}()

	return p, nil
}

func (p *Provider) Put(
	prs ...*hyperspace.Announcement,
) {
	for _, pr := range prs {
		p.peerCache.Put(pr, peerCacheTTL)
		if hasHyperspaceCapability(pr) {
			p.providerCache.Put(pr, providerCacheTTL)
		}
	}
}

func (p *Provider) handleObject(
	e *network.Envelope,
) {
	ctx := p.context

	logger := log.FromContext(ctx).With(
		log.String("method", "Provider.handleObject"),
		log.String("env.Sender", e.Sender.String()),
	)

	o := e.Payload
	switch o.Type {
	case hyperspaceAnnouncementType:
		if err := p.handleAnnouncement(ctx, o); err != nil {
			logger.Warn(
				"error handling announcement",
				log.Error(err),
			)
		}
	case hyperspaceLookupRequestType:
		v := &hyperspace.LookupRequest{}
		if err := v.UnmarshalObject(o); err != nil {
			logger.Warn(
				"error decoding lookup request",
				log.Error(err),
			)
			return
		}
		p.handlePeerLookup(ctx, v, e)
	}
}

func (p *Provider) handleAnnouncement(
	ctx context.Context,
	annObj *object.Object,
) error {
	ann := &hyperspace.Announcement{}
	if err := ann.UnmarshalObject(annObj); err != nil {
		return err
	}
	logger := log.FromContext(ctx).With(
		log.String("method", "provider.handleAnnouncement"),
		log.String("peer.publicKey", ann.ConnectionInfo.PublicKey.String()),
		log.Strings("peer.addresses", ann.ConnectionInfo.Addresses),
	)
	// TODO check if we've already received this peer, and if not forward it
	// to the other hyperspace providers
	logger.Debug("adding peer to cache")
	ok := p.peerCache.Put(ann, peerCacheTTL)
	if ok {
		p.distributeAnnouncement(annObj)
	}
	if hasHyperspaceCapability(ann) {
		p.providerCache.Put(ann, providerCacheTTL)
	}
	return nil
}

func (p *Provider) handlePeerLookup(
	ctx context.Context,
	q *hyperspace.LookupRequest,
	e *network.Envelope,
) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "provider.handlePeerLookup"),
		log.Any("q.vector", q.QueryVector),
	)

	if !e.Sender.IsEmpty() {
		logger = logger.With(
			log.String("e.sender", e.Sender.String()),
		)
	}

	if !e.Payload.Metadata.Signature.Signer.IsEmpty() {
		logger = logger.With(
			log.String(
				"o.signer",
				e.Payload.Metadata.Signature.Signer.String(),
			),
		)
	}

	promIncRequestsCounter.Inc()

	logger.Debug("handling peer lookup")

	ans := p.peerCache.Lookup(hyperspace.Bloom(q.QueryVector))

	ctx = context.New(
		context.WithParent(ctx),
	)

	promIncResponsesHistogram.Observe(float64(len(ans)))

	res := &hyperspace.LookupResponse{
		Metadata: object.Metadata{
			Owner: p.local.GetPeerKey().PublicKey(),
		},
		Nonce:         q.Nonce,
		QueryVector:   q.QueryVector,
		Announcements: ans,
	}

	pr := &peer.ConnectionInfo{
		PublicKey: e.Sender,
	}

	reso, err := res.MarshalObject()
	if err != nil {
		logger.Debug("could not marshal lookup response",
			log.Error(err),
		)
		return
	}
	err = p.network.Send(
		ctx,
		reso,
		pr.PublicKey,
		network.SendWithConnectionInfo(pr),
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

// distributeAnnouncement to all relevant hyperspace providers, this should
// only be called when we received an announcement that we did not already
// know about.
//
// NOTE: in this version of the hyperspace all providers have a complete
// picture of the network.
// TODO: consider batching up announcements somehow
func (p *Provider) distributeAnnouncement(
	annObj *object.Object,
) {
	ctx := context.New(
		context.WithParent(p.context),
	)
	logger := log.FromContext(ctx).With(
		log.String("method", "provider.distributeAnnouncement"),
	)
	n := 0
	for _, ci := range p.providerCache.List() {
		if err := p.network.Send(
			context.New(
				context.WithParent(ctx),
				context.WithTimeout(time.Second*3),
			),
			annObj,
			ci.ConnectionInfo.PublicKey,
			network.SendWithConnectionInfo(ci.ConnectionInfo),
		); err != nil {
			logger.Error(
				"error announcing self to other provider",
				log.String("provider", ci.ConnectionInfo.PublicKey.String()),
				log.Error(err),
			)
		}
		n++
	}
	logger.Info("forwarded announcement to other provider", log.Int("n", n))
}

func (p *Provider) announceSelf() {
	ctx := context.New(
		context.WithParent(p.context),
	)
	logger := log.FromContext(ctx).With(
		log.String("method", "provider.announceSelf"),
	)
	// construct an announcement
	ann := &hyperspace.Announcement{
		Version:        announceVersion,
		ConnectionInfo: p.network.GetConnectionInfo(),
		PeerCapabilities: []string{
			hyperspaceAnnouncementType,
			hyperspaceLookupRequestType,
		},
	}
	// make sure we have our own peer in the peer cache
	p.peerCache.Put(ann, 24*time.Hour)
	// and send it to all providers
	annObj, err := ann.MarshalObject()
	if err != nil {
		logger.Error("unable to marshal announcement", log.Error(err))
		return
	}
	n := 0
	for _, ci := range p.providerCache.List() {
		if err := p.network.Send(
			context.New(
				context.WithParent(ctx),
				context.WithTimeout(time.Second*3),
			),
			annObj,
			ci.ConnectionInfo.PublicKey,
			network.SendWithConnectionInfo(ci.ConnectionInfo),
		); err != nil {
			logger.Error(
				"error announcing self to other provider",
				log.String("provider", ci.ConnectionInfo.PublicKey.String()),
				log.Error(err),
			)
		}
		n++
	}
	logger.Info("announced self to other provider", log.Int("n", n))
}

func hasHyperspaceCapability(ann *hyperspace.Announcement) bool {
	for _, c := range ann.PeerCapabilities {
		if c == hyperspaceAnnouncementType {
			return true
		}
	}
	return false
}
