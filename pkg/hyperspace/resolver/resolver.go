package resolver

import (
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/patrickmn/go-cache"

	"nimona.io/internal/rand"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/hyperspace"
	"nimona.io/pkg/hyperspace/peerstore"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/log"
	"nimona.io/pkg/network"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

var (
	hyperspaceAnnouncementType   = new(hyperspace.Announcement).Type()
	hyperspaceLookupResponseType = new(hyperspace.LookupResponse).Type()

	peerCacheTTL = 1 * time.Minute
)

const (
	ErrNoPeersToAsk = errors.Error("no peers to ask")
)

//go:generate mockgen -destination=../resolvermock/resolvermock_generated.go -package=resolvermock -source=resolver.go

type (
	Resolver interface {
		Lookup(
			ctx context.Context,
			opts ...LookupOption,
		) ([]*peer.ConnectionInfo, error)
		LookupPeer(
			ctx context.Context,
			publicKey crypto.PublicKey,
		) (*peer.ConnectionInfo, error)
	}
	resolver struct {
		context                        context.Context
		network                        network.Network
		localpeer                      localpeer.LocalPeer
		peerCache                      *peerstore.PeerCache
		localPeerAnnouncementCache     *hyperspace.Announcement
		localPeerAnnouncementCacheLock sync.RWMutex
		bootstrapPeers                 []*peer.ConnectionInfo
		blocklist                      *cache.Cache
	}
	// Option for customizing a new resolver
	Option func(*resolver)
)

// New returns a new resolver
func New(
	ctx context.Context,
	netw network.Network,
	opts ...Option,
) Resolver {
	r := &resolver{
		context: ctx,
		network: netw,
		peerCache: peerstore.NewPeerCache(
			time.Minute,
			"nimona_hyperspace_resolver",
		),
		localPeerAnnouncementCacheLock: sync.RWMutex{},
		bootstrapPeers:                 []*peer.ConnectionInfo{},
		blocklist:                      cache.New(time.Second*5, time.Second*60),
	}

	for _, opt := range opts {
		opt(r)
	}

	r.localpeer = r.network.LocalPeer()

	// we are listening for all incoming object types in order to learn about
	// new peers that are talking to us so we can announce ourselves to them
	go network.HandleEnvelopeSubscription(
		r.network.Subscribe(),
		func(e *network.Envelope) error {
			go r.handleObject(e)
			return nil
		},
	)

	// register self to network
	netw.RegisterResolver(r)

	for _, p := range r.bootstrapPeers {
		r.peerCache.Put(&hyperspace.Announcement{
			ConnectionInfo: p,
		}, 0)
	}

	go func() {
		r.announceSelf()
		announceOnUpdate, cf := r.localpeer.ListenForUpdates()
		defer cf()
		announceTicker := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-announceTicker.C:
				r.announceSelf()
			case <-announceOnUpdate:
				r.announceSelf()
			}
		}
	}()

	return r
}

func (r *resolver) LookupPeer(
	ctx context.Context,
	publicKey crypto.PublicKey,
) (*peer.ConnectionInfo, error) {
	ps, err := r.Lookup(ctx, LookupByPeerKey(publicKey))
	if err != nil {
		return nil, err
	}
	if len(ps) == 0 {
		return nil, nil
	}
	return ps[0], nil
}

// Lookup finds and returns peer infos from a fingerprint
// TODO consider returning peers synchronously
func (r *resolver) Lookup(
	ctx context.Context,
	opts ...LookupOption,
) ([]*peer.ConnectionInfo, error) {
	if len(r.bootstrapPeers) == 0 {
		return nil, errors.Error("no peers to ask")
	}

	logger := log.FromContext(ctx).With(
		log.String("method", "resolver.Lookup"),
	)
	logger.Debug("looking up")

	opt := ParseLookupOptions(opts...)
	bl := hyperspace.New(opt.Lookups...)

	// send content requests to recipients
	req := &hyperspace.LookupRequest{
		Metadata: object.Metadata{
			Owner: r.localpeer.GetPrimaryPeerKey().PublicKey(),
		},
		Nonce:       rand.String(12),
		QueryVector: bl,
	}
	reqObject := req.ToObject()

	// listen for lookup responses
	resSub := r.network.Subscribe(
		network.FilterByObjectType(hyperspaceLookupResponseType),
		func(e *network.Envelope) bool {
			v := e.Payload.Data["nonce"]
			rn, ok := v.(object.String)
			return ok && string(rn) == req.Nonce
		},
	)

	go func() {
		for _, bp := range r.bootstrapPeers {
			err := r.network.Send(
				ctx,
				reqObject,
				bp.PublicKey,
				network.SendWithConnectionInfo(bp),
			)
			if err != nil {
				logger.Debug("could send request to peer", log.Error(err))
				continue
			}
			logger.Debug(
				"asked peer",
				log.String("peer", bp.PublicKey.String()),
			)
		}
	}()

	// create channel to keep peers we find
	peers := []*peer.ConnectionInfo{}
	done := make(chan struct{})

	go func() {
		for {
			e, err := resSub.Next()
			if err != nil {
				break
			}
			r := &hyperspace.LookupResponse{}
			if err := r.FromObject(e.Payload); err != nil {
				continue
			}
			// TODO verify peer?
			for _, ann := range r.Announcements {
				peers = append(peers, ann.ConnectionInfo)
			}
			close(done)
			break
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-done:
		return peers, nil
	}
}

func (r *resolver) handleObject(
	e *network.Envelope,
) {
	// attempt to recover correlation id from request id
	ctx := r.context

	logger := log.FromContext(ctx).With(
		log.String("method", "resolver.handleObject"),
		log.String("env.Sender", e.Sender.String()),
	)

	// handle payload
	o := e.Payload
	if o.Type == hyperspaceAnnouncementType {
		v := &hyperspace.Announcement{}
		if err := v.FromObject(o); err != nil {
			logger.Warn(
				"error handling announcement",
				log.Error(err),
			)
			return
		}
		r.handleAnnouncement(ctx, v)
	}
}

func (r *resolver) handleAnnouncement(
	ctx context.Context,
	p *hyperspace.Announcement,
) {
	logger := log.FromContext(ctx).With(
		log.String("method", "resolver.handleAnnouncement"),
		log.String("peer.publicKey", p.ConnectionInfo.PublicKey.String()),
		log.Strings("peer.addresses", p.ConnectionInfo.Addresses),
	)
	logger.Debug("adding peer to cache")
	r.peerCache.Put(p, peerCacheTTL)
}

func (r *resolver) announceSelf() {
	ctx := context.New(
		context.WithParent(r.context),
	)
	logger := log.FromContext(ctx).With(
		log.String("method", "resolver.announceSelf"),
	)
	n := 0
	for _, p := range r.bootstrapPeers {
		if err := r.network.Send(
			context.New(
				context.WithParent(ctx),
				context.WithTimeout(time.Second*3),
			),
			r.getLocalPeerAnnouncement().ToObject(),
			p.PublicKey,
			network.SendWithConnectionInfo(p),
		); err != nil {
			logger.Error(
				"error announcing self to bootstrap",
				log.String("peer", p.PublicKey.String()),
				log.Error(err),
			)
			continue
		}
		n++
	}
	if n == 0 {
		logger.Error(
			"failed to announce self to any bootstrap peers",
		)
		return
	}
	logger.Info(
		"announced self to bootstrap peers",
		log.Int("bootstrapPeers", n),
	)
}

func (r *resolver) getLocalPeerAnnouncement() *hyperspace.Announcement {
	r.localPeerAnnouncementCacheLock.RLock()
	lastAnnouncement := r.localPeerAnnouncementCache
	r.localPeerAnnouncementCacheLock.RUnlock()

	peerKey := r.localpeer.GetPrimaryPeerKey().PublicKey()
	certificates := r.localpeer.GetCertificates()
	cids := r.localpeer.GetCIDs()
	contentTypes := r.localpeer.GetContentTypes()
	addresses := r.localpeer.GetAddresses()
	relays := r.localpeer.GetRelays()

	// gather up peer key, certificates, content ids and types
	hs := contentTypes
	hs = append(hs, peerKey.String())
	for _, c := range cids {
		hs = append(hs, c.String())
	}
	for _, c := range certificates {
		if !c.Metadata.Signature.IsEmpty() {
			hs = append(hs, c.Metadata.Signature.Signer.String())
		}
	}
	vec := hyperspace.New(hs...)

	if lastAnnouncement != nil &&
		cmp.Equal(lastAnnouncement.ConnectionInfo.Addresses, addresses) &&
		cmp.Equal(lastAnnouncement.PeerVector, vec) {
		return lastAnnouncement
	}

	localPeerAnnouncementCache := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: peerKey,
		},
		Version: time.Now().Unix(),
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: peerKey,
			Addresses: addresses,
			Relays:    relays,
		},
		PeerVector:       vec,
		PeerCapabilities: contentTypes,
	}

	r.localPeerAnnouncementCacheLock.Lock()
	r.localPeerAnnouncementCache = localPeerAnnouncementCache
	r.localPeerAnnouncementCacheLock.Unlock()

	return localPeerAnnouncementCache
}
