package provider

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"nimona.io/internal/net"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace"
	"nimona.io/pkg/hyperspace/peerstore"
	"nimona.io/pkg/log"
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
		context             context.Context
		network             net.Network
		announcementVersion int64
		peerKey             crypto.PrivateKey
		peerCache           *peerstore.PeerCache
		providerCache       *peerstore.PeerCache
	}
)

func New(
	ctx context.Context,
	network net.Network,
	peerKey crypto.PrivateKey,
	bootstrapProviders []*peer.ConnectionInfo,
) (*Provider, error) {
	p := &Provider{
		context:             ctx,
		network:             network,
		announcementVersion: time.Now().Unix(),
		peerKey:             peerKey,
		peerCache: peerstore.NewPeerCache(
			peerCacheGC,
			"nimona_hyperspace_provider_peers",
		),
		providerCache: peerstore.NewPeerCache(
			providerCacheGC,
			"nimona_hyperspace_provider_bootstraps",
		),
	}

	// we are listening for all incoming object types in order to learn about
	// new peers that are talking to us so we can announce ourselves to them
	p.network.RegisterConnectionHandler(
		func(c net.Connection) {
			go func() {
				or := c.Read(ctx)
				for {
					o, err := or.Read()
					if err != nil {
						return
					}
					go p.handleObject(c.RemotePeerKey(), o)
				}
			}()
		},
	)

	for _, ci := range bootstrapProviders {
		p.providerCache.Put(
			&hyperspace.Announcement{
				ConnectionInfo: ci,
				PeerCapabilities: []string{
					hyperspace.AnnouncementType,
					hyperspace.LookupRequestType,
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
	s crypto.PublicKey,
	o *object.Object,
) {
	ctx := p.context

	logger := log.FromContext(ctx).With(
		log.String("method", "Provider.handleObject"),
		log.String("env.Sender", s.String()),
	)

	switch o.Type {
	case hyperspace.AnnouncementType:
		if err := p.handleAnnouncement(ctx, o); err != nil {
			logger.Warn(
				"error handling announcement",
				log.Error(err),
			)
		}
	case hyperspace.LookupRequestType:
		v := &hyperspace.LookupRequest{}
		if err := object.Unmarshal(o, v); err != nil {
			logger.Warn(
				"error decoding lookup request",
				log.Error(err),
			)
			return
		}
		p.handlePeerLookup(ctx, v, s, o)
	}
}

func (p *Provider) handleAnnouncement(
	ctx context.Context,
	annObj *object.Object,
) error {
	ann := &hyperspace.Announcement{}
	if err := object.Unmarshal(annObj, ann); err != nil {
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
	s crypto.PublicKey,
	o *object.Object,
) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "provider.handlePeerLookup"),
		log.Any("q.vector", q.QueryVector),
	)

	if !s.IsEmpty() {
		logger = logger.With(
			log.String("sender", s.String()),
		)
	}

	if !o.Metadata.Signature.Signer.IsEmpty() {
		logger = logger.With(
			log.String(
				"o.signer",
				o.Metadata.Signature.Signer.String(),
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
			Owner: p.peerKey.PublicKey().DID(),
		},
		Nonce:         q.Nonce,
		QueryVector:   q.QueryVector,
		Announcements: ans,
	}

	pr := &peer.ConnectionInfo{
		PublicKey: s,
	}

	reso, err := object.Marshal(res)
	if err != nil {
		logger.Debug("could not marshal lookup response",
			log.Error(err),
		)
		return
	}

	pc, err := p.network.Dial(ctx, pr)
	if err != nil {
		logger.Debug("could not dial peer",
			log.Error(err),
		)
		return
	}

	err = pc.Write(ctx, reso)
	if err != nil {
		logger.Debug("could not write lookup response", log.Error(err))
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
		pc, err := p.network.Dial(ctx, ci.ConnectionInfo)
		if err != nil {
			logger.Debug("could not dial peer", log.Error(err))
			return
		}
		if err := pc.Write(
			context.New(
				context.WithParent(ctx),
				context.WithTimeout(time.Second*3),
			),
			annObj,
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
		Version: announceVersion,
		ConnectionInfo: &peer.ConnectionInfo{
			Version:   p.announcementVersion,
			PublicKey: p.peerKey.PublicKey(),
			Addresses: p.network.Addresses(),
		},
		PeerCapabilities: []string{
			hyperspace.AnnouncementType,
			hyperspace.LookupRequestType,
		},
	}
	// make sure we have our own peer in the peer cache
	p.peerCache.Put(ann, 24*time.Hour)
	// and send it to all providers
	annObj, err := object.Marshal(ann)
	if err != nil {
		logger.Error("unable to marshal announcement", log.Error(err))
		return
	}
	n := 0
	for _, ci := range p.providerCache.List() {
		pc, err := p.network.Dial(ctx, ci.ConnectionInfo)
		if err != nil {
			logger.Error("unable to dial provider", log.Error(err))
			return
		}
		if err := pc.Write(
			context.New(
				context.WithParent(ctx),
				context.WithTimeout(time.Second*3),
			),
			annObj,
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
		if c == hyperspace.AnnouncementType {
			return true
		}
	}
	return false
}
