package hyperspace

import (
	"context"
	"time"

	"go.uber.org/zap"

	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net"
	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object"
	"nimona.io/pkg/object/exchange"
)

var (
	typePeerInfoRequest = peer.PeerInfoRequest{}.GetType()
	typePeerInfo        = peer.PeerInfo{}.GetType()
)

// Discoverer hyperspace
type Discoverer struct {
	store    *Store
	key      *crypto.Key
	net      net.Network
	exchange exchange.Exchange
}

// NewDiscoverer returns a new hyperspace discoverer
func NewDiscoverer(key *crypto.Key, network net.Network, exc exchange.Exchange,
	bootstrapAddresses []string) (*Discoverer, error) {

	r := &Discoverer{
		store:    NewStore(),
		key:      key,
		net:      network,
		exchange: exc,
	}

	exc.Handle("/peer**", r.handleBlock)

	r.store.Add(network.GetPeerInfo())
	r.bootstrap(bootstrapAddresses)

	return r, nil
}

// Discover finds and returns the closest peers to a query
func (r *Discoverer) Discover(q *peer.PeerInfoRequest) ([]*peer.PeerInfo, error) {
	ctx := context.Background()
	eps := r.store.FindExact(q)
	go r.LookupPeerInfo(ctx, q)
	// TODO(geoah) use dht-like queries instead of a delay
	time.Sleep(time.Second)
	cps := r.store.FindClosest(q)
	ps := append(eps, cps...)
	return ps, nil
}

func (r *Discoverer) handleBlock(o *object.Object) error {
	switch o.GetType() {
	case typePeerInfoRequest:
		v := &peer.PeerInfoRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerInfoRequest(v)
	case typePeerInfo:
		v := &peer.PeerInfo{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerInfo(v)
	}
	return nil
}

func (r *Discoverer) handlePeerInfo(p *peer.PeerInfo) {
	r.store.Add(p)
}

func (r *Discoverer) handlePeerInfoRequest(q *peer.PeerInfoRequest) {
	ctx := context.Background()
	logger := log.Logger(ctx)
	eps := r.store.FindExact(q)
	cps := r.store.FindClosest(q)
	ps := append(eps, cps...)
	for _, p := range ps {
		addr := "peer:" + q.RequesterSignerKey.HashBase58()
		if err := r.exchange.Send(ctx, p.ToObject(), addr); err != nil {
			logger.Debug("handleProviderRequest could not send block", zap.Error(err))
		}
	}
}

// LookupPeerInfo does a network lookup given a query
func (r *Discoverer) LookupPeerInfo(ctx context.Context, q *peer.PeerInfoRequest) error {
	o := q.ToObject()
	if err := crypto.Sign(o, r.key); err != nil {
		return err
	}
	ps := r.store.FindClosest(q)
	for _, p := range ps {
		r.exchange.Send(ctx, o, "peer:"+p.SignerKey.HashBase58())
	}
	return nil
}

func (r *Discoverer) bootstrap(bootstrapAddresses []string) error {
	ctx := context.Background()
	logger := log.Logger(ctx)
	for _, addr := range bootstrapAddresses {
		q := &peer.PeerInfoRequest{
			SignerKeyHash: r.key.GetPublicKey().HashBase58(),
		}
		o := q.ToObject()
		if err := crypto.Sign(o, r.key); err != nil {
			continue
		}
		if err := r.exchange.Send(ctx, o, addr); err != nil {
			logger.Debug("bootstrap could not send request", zap.Error(err))
		}
		if err := r.exchange.Send(ctx, r.net.GetPeerInfo().ToObject(), addr); err != nil {
			logger.Debug("bootstrap could not send self", zap.Error(err))
		}
	}
	return nil
}
