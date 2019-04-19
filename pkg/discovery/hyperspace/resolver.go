package hyperspace

import (
	"context"
	"time"

	"go.uber.org/zap"

	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net"
	"nimona.io/pkg/net/peer"
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
	local    *net.LocalInfo
}

// NewDiscoverer returns a new hyperspace discoverer
func NewDiscoverer(key *crypto.Key, network net.Network, exc exchange.Exchange,
	local *net.LocalInfo, bootstrapAddresses []string) (*Discoverer, error) {

	r := &Discoverer{
		store:    NewStore(),
		key:      key,
		net:      network,
		local:    local,
		exchange: exc,
	}

	exc.Handle("/peer**", r.handleObject)

	r.store.Add(local.GetPeerInfo())
	r.bootstrap(bootstrapAddresses)

	return r, nil
}

// Discover finds and returns the closest peers to a query
func (r *Discoverer) Discover(q *peer.PeerInfoRequest) ([]*peer.PeerInfo, error) {
	ctx := context.Background()
	go r.LookupPeerInfo(ctx, q)
	// TODO(geoah) use dht-like queries instead of a delay
	time.Sleep(time.Second)
	// cps := r.store.FindClosest(q)
	// ps := append(eps, cps...)
	eps := r.store.FindExact(q)
	return eps, nil
}

func (r *Discoverer) handleObject(e *exchange.Envelope) error {
	s := e.Sender
	o := e.Payload
	switch o.GetType() {
	case typePeerInfoRequest:
		v := &peer.PeerInfoRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerInfoRequest(s, v)
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
	logger := log.DefaultLogger.With(
		zap.String("method", "resolver/handlePeerInfo"),
		zap.String("peerinfo._hash", p.HashBase58()),
		zap.Strings("peerinfo.addresses", p.Addresses),
	)
	logger.Debug("adding peerinfo to store")
	r.store.Add(p)
}

func (r *Discoverer) handlePeerInfoRequest(s *crypto.Key, q *peer.PeerInfoRequest) {
	ctx := context.Background()
	logger := log.Logger(ctx)
	eps := r.store.FindExact(q)
	cps := r.store.FindClosest(q)
	ps := append(eps, cps...)
	for _, p := range ps {
		addr := "peer:" + s.HashBase58()
		if err := r.exchange.Send(ctx, p.ToObject(), addr); err != nil {
			logger.Debug("handleProviderRequest could not send object", zap.Error(err))
		}
	}
}

// LookupPeerInfo does a network lookup given a query
func (r *Discoverer) LookupPeerInfo(ctx context.Context, q *peer.PeerInfoRequest) error {
	logger := log.DefaultLogger.With(
		zap.String("method", "resolver/handlePeerInfo"),
		zap.Strings("query.contentIDs", q.ContentIDs),
	)
	o := q.ToObject()
	ps := r.store.FindClosest(q)
	logger.Debug("found closest peers", zap.Int("n", len(ps)))
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
		if err := r.exchange.Send(ctx, r.local.GetPeerInfo().ToObject(), addr); err != nil {
			logger.Debug("bootstrap could not send self", zap.Error(err))
		}
	}
	return nil
}
