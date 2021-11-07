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
	"nimona.io/pkg/log"
	"nimona.io/pkg/net"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/sqlobjectstore"
	"nimona.io/pkg/tilde"
)

const (
	ErrNoPeersToAsk = errors.Error("no peers to ask")

	peerCacheTTL = 1 * time.Minute
)

//go:generate mockgen -destination=../resolvermock/resolvermock_generated.go -package=resolvermock -source=resolver.go
//go:generate genny -in=$GENERATORS/synclist/synclist.go -out=hashes_generated.go -imp=nimona.io/pkg/object -pkg=resolver gen "KeyType=tilde.Digest"

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
		peerKey                        crypto.PrivateKey
		context                        context.Context
		network                        net.Network
		peerCache                      *peerstore.PeerCache
		localPeerAnnouncementCache     *hyperspace.Announcement
		localPeerAnnouncementCacheLock sync.RWMutex
		bootstrapPeers                 []*peer.ConnectionInfo
		blocklist                      *cache.Cache
		hashes                         *TildeDigestSyncList
	}
	// Option for customizing a new resolver
	Option func(*resolver)
)

// New returns a new resolver.
// Object store is currently optional.
func New(
	ctx context.Context,
	network net.Network,
	peerKey crypto.PrivateKey,
	str *sqlobjectstore.Store,
	opts ...Option,
) Resolver {
	r := &resolver{
		context: ctx,
		peerKey: peerKey,
		network: network,
		peerCache: peerstore.NewPeerCache(
			time.Minute,
			"nimona_hyperspace_resolver",
		),
		localPeerAnnouncementCacheLock: sync.RWMutex{},
		bootstrapPeers:                 []*peer.ConnectionInfo{},
		blocklist:                      cache.New(time.Second*5, time.Second*60),
		hashes:                         &TildeDigestSyncList{},
	}

	for _, opt := range opts {
		opt(r)
	}

	// we are listening for all incoming object types in order to learn about
	// new peers that are talking to us so we can announce ourselves to them
	go r.network.RegisterConnectionHandler(
		func(c net.Connection) {
			go func() {
				or := c.Read(ctx)
				for {
					o, err := or.Read()
					if err != nil {
						return
					}
					r.handleObject(c.RemotePeerKey(), o)
				}
			}()
		},
	)

	// go through all existing objects and add them as well
	if str != nil {
		if hashes, err := str.ListHashes(); err == nil {
			for _, hash := range hashes {
				r.hashes.Put(hash)
			}
		}
	}

	for _, p := range r.bootstrapPeers {
		r.peerCache.Put(&hyperspace.Announcement{
			ConnectionInfo: p,
		}, 0)
	}

	go func() {
		// announce on startup
		r.announceSelf()

		// TODO(geoah): fix identity
		// subscribe to local peer updates
		// lpSub, lpCf := r.localpeer.ListenForUpdates()
		// defer lpCf()

		// subscribe to object updates
		strSub := make(<-chan sqlobjectstore.Event)
		strCf := func() {}
		if str != nil {
			strSub, strCf = str.ListenForUpdates()
		}
		defer strCf()

		// or every 30 seconds
		announceTicker := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-announceTicker.C:
				r.announceSelf()
				// TODO(geoah): fix identity
				// case <-lpSub:
				// 	r.announceSelf()
			case event := <-strSub:
				switch event.Action {
				case sqlobjectstore.ObjectInserted:
					r.hashes.Put(event.ObjectHash)
					r.announceSelf()
				case sqlobjectstore.ObjectRemoved:
					r.hashes.Delete(event.ObjectHash)
					r.announceSelf()
				}
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
			Owner: r.peerKey.PublicKey().DID(),
		},
		Nonce:       rand.String(12),
		QueryVector: bl,
	}
	reqObject, err := object.Marshal(req)
	if err != nil {
		return nil, err
	}

	responses := make(chan *object.Object)

	go func() {
		for _, bp := range r.bootstrapPeers {
			conn, err := r.network.Dial(ctx, bp)
			if err != nil {
				logger.Debug("failed to dial peer", log.Error(err))
				continue
			}
			// read all objects from the connection
			go func() {
				reader := conn.Read(ctx)
				for {
					o, err := reader.Read()
					if err != nil {
						return
					}
					// find any responses
					v := o.Data["nonce"]
					rn, ok := v.(tilde.String)
					if ok && string(rn) == req.Nonce {
						// and write them to the channel
						responses <- o
					}
				}
			}()
			err = conn.Write(
				ctx,
				reqObject,
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
			o := <-responses
			r := &hyperspace.LookupResponse{}
			if err := object.Unmarshal(o, r); err != nil {
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
	sender crypto.PublicKey,
	o *object.Object,
) {
	// attempt to recover correlation id from request id
	ctx := r.context

	logger := log.FromContext(ctx).With(
		log.String("method", "resolver.handleObject"),
		log.String("env.Sender", sender.String()),
	)

	// handle payload
	// o := e.Payload
	if o.Type == hyperspace.AnnouncementType {
		v := &hyperspace.Announcement{}
		if err := object.Unmarshal(o, v); err != nil {
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
	anno, err := object.Marshal(r.getLocalPeerAnnouncement())
	if err != nil {
		logger.Error("error marshaling announcement", log.Error(err))
		return
	}
	for _, p := range r.bootstrapPeers {
		conn, err := r.network.Dial(
			context.New(
				context.WithParent(ctx),
				context.WithTimeout(time.Second*3),
			),
			p,
		)
		if err != nil {
			logger.Debug("failed to dial peer", log.Error(err))
			continue
		}
		err = conn.Write(
			ctx,
			anno,
		)
		if err != nil {
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

	peerKey := r.peerKey.PublicKey()
	hashes := r.hashes.List()
	addresses := r.network.Addresses()
	// TODO support relays
	// relays := r.network.GetRelays()

	// gather up peer key, certificates, content ids and types
	hs := []string{peerKey.String()}
	for _, c := range hashes {
		hs = append(hs, c.String())
	}
	// TODO(geoah): fix identity
	// if c := r.localpeer.GetPeerCertificate(); c != nil {
	// 	if !c.Metadata.Signature.IsEmpty() {
	// 		hs = append(hs, c.Metadata.Signature.Signer.String())
	// 	}
	// }
	vec := hyperspace.New(hs...)

	if lastAnnouncement != nil &&
		cmp.Equal(lastAnnouncement.ConnectionInfo.Addresses, addresses) &&
		cmp.Equal(lastAnnouncement.PeerVector, vec) {
		return lastAnnouncement
	}

	localPeerAnnouncementCache := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: peerKey.DID(),
		},
		Version: time.Now().Unix(),
		ConnectionInfo: &peer.ConnectionInfo{
			PublicKey: peerKey,
			Addresses: addresses,
			// Relays:    relays,
		},
		PeerVector: vec,
		// TODO add capabilities
	}

	r.localPeerAnnouncementCacheLock.Lock()
	r.localPeerAnnouncementCache = localPeerAnnouncementCache
	r.localPeerAnnouncementCacheLock.Unlock()

	return localPeerAnnouncementCache
}
