package hyperspace

import (
	"context"
	"time"

	"go.uber.org/zap"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/encoding"
	"nimona.io/pkg/log"
	"nimona.io/pkg/net"
	"nimona.io/pkg/peers"
)

var (
	typePeerInfoRequest = peers.PeerInfoRequest{}.GetType()
	typePeerInfo        = peers.PeerInfo{}.GetType()
)

// Resolver hyperspace
type Resolver struct {
	store    *Store
	key      *crypto.Key
	net      net.Network
	exchange net.Exchange
}

// NewResolver returns a new hyperspace resolver
func NewResolver(key *crypto.Key, network net.Network, exchange net.Exchange,
	bootstrapAddresses []string) (*Resolver, error) {

	r := &Resolver{
		store:    NewStore(),
		key:      key,
		net:      network,
		exchange: exchange,
	}

	exchange.Handle("/peer**", r.handleBlock)

	r.store.Add(network.GetPeerInfo())
	r.bootstrap(bootstrapAddresses)

	return r, nil
}

// Resolve finds and returns the closest peers to a query
func (r *Resolver) Resolve(q *peers.PeerInfoRequest) ([]*peers.PeerInfo, error) {
	ctx := context.Background()
	eps := r.store.FindExact(q)
	go r.LookupPeerInfo(ctx, q)
	// TODO(geoah) use dht-like queries instead of a delay
	time.Sleep(time.Second)
	cps := r.store.FindClosest(q)
	ps := append(eps, cps...)
	return ps, nil
}

func (r *Resolver) handleBlock(o *encoding.Object) error {
	switch o.GetType() {
	case typePeerInfoRequest:
		v := &peers.PeerInfoRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerInfoRequest(v)
	case typePeerInfo:
		v := &peers.PeerInfo{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerInfo(v)
	}
	return nil
}

func (r *Resolver) handlePeerInfo(p *peers.PeerInfo) {
	r.store.Add(p)
}

func (r *Resolver) handlePeerInfoRequest(q *peers.PeerInfoRequest) {
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
func (r *Resolver) LookupPeerInfo(ctx context.Context, q *peers.PeerInfoRequest) error {
	o := q.ToObject()
	if err := crypto.Sign(o, r.key); err != nil {
		return err
	}
	ps := r.store.FindClosest(q)
	for _, p := range ps {
		r.exchange.Send(ctx, o, "peers:"+p.SignerKey.HashBase58())
	}
	return nil
}

func (r *Resolver) bootstrap(bootstrapAddresses []string) error {
	ctx := context.Background()
	logger := log.Logger(ctx)
	for _, addr := range bootstrapAddresses {
		q := &peers.PeerInfoRequest{
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
