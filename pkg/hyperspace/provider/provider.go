package provider

import (
	"time"

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
	peerType              = new(peer.Peer).Type()
	peerLookupRequestType = new(peer.LookupRequest).Type()
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
	c := peerstore.NewPeerCache(peerCacheGC)
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
	prs ...*peer.Peer,
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
	case peerType:
		v := &peer.Peer{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		p.handlePeer(ctx, v)
	case peerLookupRequestType:
		v := &peer.LookupRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		p.handlePeerLookup(ctx, v, e)
		return nil
	}

	return nil
}

func (p *Provider) handlePeer(
	ctx context.Context,
	incPeer *peer.Peer,
) {
	logger := log.FromContext(ctx).With(
		log.String("method", "provider.handlePeer"),
		log.String("peer.publicKey", incPeer.PublicKey().String()),
		log.Strings("peer.addresses", incPeer.Addresses),
	)
	// TODO check if we've already received this peer, and if not forward it
	// to the other hyperspace providers
	logger.Debug("adding peer to cache")
	p.peerCache.Put(incPeer, peerCacheTTL)
}

func (p *Provider) handlePeerLookup(
	ctx context.Context,
	q *peer.LookupRequest,
	e *network.Envelope,
) {
	ctx = context.FromContext(ctx)
	logger := log.FromContext(ctx).With(
		log.String("method", "provider.handlePeerLookup"),
		log.String("e.sender", e.Sender.String()),
		log.Any("q.vector", q.QueryVector),
		log.Any("o.signer", e.Payload.Metadata.Signature.Signer),
	)

	logger.Debug("handling peer lookup")

	ps := p.peerCache.Lookup(hyperspace.Bloom(q.QueryVector))

	ctx = context.New(
		context.WithParent(ctx),
	)

	res := &peer.LookupResponse{
		Metadata: object.Metadata{
			Owner: p.local.GetPrimaryPeerKey().PublicKey(),
		},
		Nonce:       q.Nonce,
		QueryVector: q.QueryVector,
		Peers:       ps,
	}

	pr := &peer.Peer{
		Metadata: object.Metadata{
			Owner: e.Sender,
		},
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
		log.Int("n", len(ps)),
	).Debug("handling done, sent n peers")
}
