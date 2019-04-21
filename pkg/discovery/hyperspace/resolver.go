package hyperspace

import (
	"runtime"
	"time"

	"go.uber.org/zap"

	"nimona.io/internal/context"
	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net"
	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object/exchange"
)

// Discoverer hyperspace
type Discoverer struct {
	store    *Store
	net      net.Network
	exchange exchange.Exchange
	local    *net.LocalInfo
}

// NewDiscoverer returns a new hyperspace discoverer
func NewDiscoverer(
	network net.Network,
	exc exchange.Exchange,
	local *net.LocalInfo,
	bootstrapAddresses []string,
) (*Discoverer, error) {
	r := &Discoverer{
		store:    NewStore(),
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
func (r *Discoverer) Discover(
	ctx context.Context,
	q *peer.PeerInfoRequest,
) ([]*peer.PeerInfo, error) {
	go r.LookupPeerInfo(ctx, q)
	// TODO(geoah) use dht-like queries instead of a delay
	time.Sleep(time.Second)
	// cps := r.store.FindClosest(q)
	// ps := append(eps, cps...)
	eps := r.store.FindExact(q)
	return eps, nil
}

func (r *Discoverer) handleObject(e *exchange.Envelope) error {
	o := e.Payload
	switch o.GetType() {
	case peer.PeerInfoRequestType:
		v := &peer.PeerInfoRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerInfoRequest(v, e)
	case peer.PeerInfoType:
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

func (r *Discoverer) handlePeerInfoRequest(q *peer.PeerInfoRequest, e *exchange.Envelope) {
	ctx := context.Background()
	logger := log.Logger(ctx)
	eps := r.store.FindExact(q)
	cps := r.store.FindClosest(q)
	ps := append(eps, cps...)
	opts := []exchange.Option{
		exchange.WithLocalDiscoveryOnly(),
		exchange.AsResponse(e.RequestID),
	}
	for _, p := range ps {
		addr := "peer:" + e.Sender.HashBase58()
		err := r.exchange.Send(ctx, p.ToObject(), addr, opts...)
		if err != nil {
			logger.Debug("handleProviderRequest could not send object",
				zap.Error(err),
			)
		}
	}
}

// LookupPeerInfo does a network lookup given a query
func (r *Discoverer) LookupPeerInfo(ctx context.Context, q *peer.PeerInfoRequest) error {
	logger := log.Logger(ctx).With(
		zap.String("method", "resolver/LookupPeerInfo"),
		zap.Strings("query.contentIDs", q.ContentIDs),
		zap.Strings("query.contentTypes", q.ContentTypes),
		zap.String("query.signerKeyHash", q.SignerKeyHash),
		zap.String("query.authorityKeyHash", q.AuthorityKeyHash),
	)
	o := q.ToObject()
	ps := r.store.FindClosest(q)
	out := make(chan *exchange.Envelope, 10)
	opts := []exchange.Option{
		exchange.WithLocalDiscoveryOnly(),
		exchange.WithResponse("", out),
	}
	logger.Debug("found closest peers", zap.Int("n", len(ps)))
	for _, p := range ps {
		r.exchange.Send(ctx, o, "peer:"+p.SignerKey.HashBase58(), opts...)
	}
	// TODO(geoah) better timeout
	t := time.NewTicker(time.Second * 5)
	for {
		select {
		case <-t.C:
			return nil

		case <-ctx.Done():
			return nil

		case res := <-out:
			logger.Debug("got response",
				zap.String("res.type", res.Payload.GetType()),
				zap.String("res.sender", res.Sender.HashBase58()),
			)
			r.handleObject(res)
		}
	}
	return nil
}

func (r *Discoverer) bootstrap(bootstrapAddresses []string) error {
	ctx := context.Background()
	logger := log.Logger(ctx)
	key := r.local.GetPeerKey()
	opts := []exchange.Option{
		exchange.WithLocalDiscoveryOnly(),
	}
	for _, addr := range bootstrapAddresses {
		q := &peer.PeerInfoRequest{
			SignerKeyHash: key.GetPublicKey().HashBase58(),
		}
		o := q.ToObject()
		err := crypto.Sign(o, key)
		if err != nil {
			continue
		}
		err = r.exchange.Send(ctx, o, addr, opts...)
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

func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	return frame
}
