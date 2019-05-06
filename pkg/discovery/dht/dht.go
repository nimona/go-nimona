package dht

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"

	"nimona.io/internal/log"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net"
	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object/exchange"
)

var (
	ErrNotFound = errors.New("not found")
)

const (
	closestPeersToReturn = 8
	maxQueryTime         = 2 * time.Second
)

var (
	typePeerInfoRequest  = PeerInfoRequest{}.GetType()
	typePeerInfoResponse = PeerInfoResponse{}.GetType()
	typeProviderRequest  = ProviderRequest{}.GetType()
	typeProviderResponse = ProviderResponse{}.GetType()
	typePeerInfo         = peer.PeerInfo{}.GetType()
)

// DHT is the struct that implements the dht protocol
type DHT struct {
	peerID         string
	store          *Store
	peerStore      *peer.PeerInfoCollection
	net            net.Network
	exchange       exchange.Exchange
	queries        sync.Map
	key            *crypto.PrivateKey
	local          *net.LocalInfo
	refreshBuckets bool
}

// NewDHT returns a new DHT from a exchange and peer manager
func NewDHT(key *crypto.PrivateKey, network net.Network, exchange exchange.Exchange,
	local *net.LocalInfo, bootstrapAddresses []string) (*DHT, error) {

	// create new kv store for storing providers
	store, _ := newStore()

	// Create DHT node
	r := &DHT{
		net:       network,
		store:     store,
		exchange:  exchange,
		queries:   sync.Map{},
		key:       key,
		peerStore: &peer.PeerInfoCollection{},
		local:     local,
	}

	exchange.Handle("nimona.io/dht/**", r.handleObject)
	exchange.Handle("/peer", r.handleObject)

	// connect to the bootstrap addresses to get their peer infos
	for _, addr := range bootstrapAddresses {
		ctx := context.Background()
		req := &PeerInfoRequest{
			RequestID: net.RandStringBytesMaskImprSrc(8),
			PeerID:    key.PublicKey.Fingerprint(),
		}
		so := req.ToObject()
		if err := crypto.Sign(so, key); err != nil {
			// TODO log error
			continue
		}
		if err := exchange.Send(ctx, so, addr); err != nil {
			log.Logger(ctx).Warn("could not send to bootstrap", zap.String("addr", addr), zap.Error(err))
		}
	}

	// start refresh process
	// TODO(geoah) enable or replace
	// go r.refresh()

	return r, nil
}

func (r *DHT) refresh() {
	ctx := context.Background()
	logger := log.Logger(ctx)
	// TODO this will be replaced when we introduce bucketing
	// TODO our init process is a bit messed up and addressBook doesn't know
	// about the peer's protocols instantly
	for {
		peerInfo := r.local.GetPeerInfo()
		if len(peerInfo.Addresses) == 0 {
			time.Sleep(time.Second * 10)
			continue
		}

		closestPeers, err := r.FindPeersClosestTo(peerInfo.Fingerprint(), closestPeersToReturn)
		if err != nil {
			logger.Warn("refresh could not get peers ids", zap.Error(err))
			time.Sleep(time.Second * 10)
			continue
		}

		// announce our peer info to the closest peers
		for _, closestPeer := range closestPeers {
			if err := r.exchange.Send(ctx, peerInfo.ToObject(), closestPeer.Address()); err != nil {
				logger.Debug("refresh could not announce", zap.Error(err), zap.String("peerID", closestPeer.Fingerprint()))
			}
		}

		// HACK lookup our own peer info just so we can populate our peer table
		r.GetPeerInfo(ctx, peerInfo.Fingerprint())

		// sleep for a bit
		time.Sleep(time.Second * 30)
	}
}

func (r *DHT) handleObject(e *exchange.Envelope) error {
	o := e.Payload
	switch o.GetType() {
	case typePeerInfoRequest:
		v := &PeerInfoRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerInfoRequest(v, e.Sender)
	case typePeerInfoResponse:
		v := &PeerInfoResponse{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerInfoResponse(v)
	case typeProviderRequest:
		v := &ProviderRequest{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handleProviderRequest(v, e.Sender)
	case typeProviderResponse:
		v := &ProviderResponse{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handleProviderResponse(v)
	case typePeerInfo:
		v := &peer.PeerInfo{}
		if err := v.FromObject(o); err != nil {
			return err
		}
		r.handlePeerInfo(v)
	default:
		return nil
	}
	return nil
}

func (r *DHT) handlePeerInfo(payload *peer.PeerInfo) {
	if err := r.peerStore.Put(payload); err != nil {
		log.Logger(context.Background()).Error("could not handle peer info", zap.Error(err))
	}
}

func (r *DHT) handlePeerInfoRequest(payload *PeerInfoRequest, sender *crypto.PublicKey) {
	ctx := context.Background()
	logger := log.Logger(ctx)

	peerInfo, _ := r.peerStore.Get(payload.PeerID)
	// TODO handle and log error

	if peerInfo == nil {
		// peerInfo, _ = r.net.Discoverer().Discover(payload.PeerID, net.Local())
		// TODO handle and log error
	}

	closestPeerInfos, err := r.FindPeersClosestTo(payload.PeerID, closestPeersToReturn)
	if err != nil {
		logger.Debug("could not get providers from local store", zap.Error(err))
		// TODO handle and log error
	}

	resp := &PeerInfoResponse{
		RequestID:    payload.RequestID,
		PeerInfo:     peerInfo,
		ClosestPeers: closestPeerInfos,
	}

	so := resp.ToObject()
	if err := crypto.Sign(so, r.key); err != nil {
		// TODO log error
		return
	}
	addr := "peer:" + sender.Fingerprint()
	if err := r.exchange.Send(ctx, so, addr); err != nil {
		logger.Debug("handleProviderRequest could not send object", zap.Error(err))
		return
	}
}

func (r *DHT) handlePeerInfoResponse(payload *PeerInfoResponse) {
	ctx := context.Background()
	logger := log.Logger(ctx)
	for _, pi := range payload.ClosestPeers {
		if err := r.peerStore.Put(pi); err != nil {
			logger.Error("could not handle closest peer from peerinfo response", zap.Error(err))
		}
	}

	if payload.PeerInfo != nil {
		if err := r.peerStore.Put(payload.PeerInfo); err != nil {
			logger.Error("could not handle peer info from peerinfo response", zap.Error(err))
		}
	}

	rID := payload.RequestID
	if rID == "" {
		return
	}

	q, exists := r.queries.Load(rID)
	if !exists {
		return
	}

	q.(*query).incomingPayloads <- payload.PeerInfo
}

func (r *DHT) handleProviderRequest(payload *ProviderRequest, sender *crypto.PublicKey) {
	ctx := context.Background()
	logger := log.Logger(ctx)

	providers, err := r.store.GetProviders(payload.Key)
	if err != nil {
		logger.Debug("could not get providers from local store", zap.Error(err))
		// TODO handle and log error
	}

	closestPeerInfos, err := r.FindPeersClosestTo(payload.Key, closestPeersToReturn)
	if err != nil {
		logger.Debug("could not get providers from local store", zap.Error(err))
		// TODO handle and log error
	}

	resp := &ProviderResponse{
		RequestID:    payload.RequestID,
		Providers:    providers,
		ClosestPeers: closestPeerInfos,
	}

	addr := "peer:" + sender.Fingerprint()
	so := resp.ToObject()
	if err := crypto.Sign(so, r.key); err != nil {
		// TODO log error
		return
	}
	if err := r.exchange.Send(ctx, so, addr); err != nil {
		logger.Warn("handleProviderRequest could not send object", zap.Error(err))
		return
	}
}

func (r *DHT) handleProviderResponse(payload *ProviderResponse) {
	ctx := context.Background()
	logger := log.Logger(ctx)

	for _, provider := range payload.Providers {
		if err := r.store.PutProvider(provider); err != nil {
			logger.Debug("could not store provider", zap.Error(err))
			// TODO handle error
		}
	}

	rID := payload.RequestID
	if rID == "" {
		return
	}

	q, exists := r.queries.Load(rID)
	if !exists {
		return
	}

	q.(*query).incomingPayloads <- payload
}

// FindPeersClosestTo returns an array of n peers closest to the given key by xor distance
func (r *DHT) FindPeersClosestTo(tk string, n int) ([]*peer.PeerInfo, error) {
	// place to hold the results
	rks := []*peer.PeerInfo{}

	htk := hash(tk)

	peerInfos, _ := r.peerStore.All()
	// slice to hold the distances
	dists := []distEntry{}
	for _, peerInfo := range peerInfos {
		peerInfoThumbprint := peerInfo.Fingerprint()
		// calculate distance
		de := distEntry{
			key:      peerInfoThumbprint,
			dist:     xor([]byte(htk), []byte(hash(peerInfoThumbprint))),
			peerInfo: peerInfo,
		}
		exists := false
		for _, ee := range dists {
			if ee.key == peerInfoThumbprint {
				exists = true
				break
			}
		}
		if !exists {
			dists = append(dists, de)
		}
	}

	// sort the distances
	sort.Slice(dists, func(i, j int) bool {
		return lessIntArr(dists[i].dist, dists[j].dist)
	})

	if n > len(dists) {
		n = len(dists)
	}

	// append n the first n number of keys
	for _, de := range dists {
		rks = append(rks, de.peerInfo)
		n--
		if n == 0 {
			break
		}
	}

	return rks, nil
}

// Discover returns a peer's info from their id
func (r *DHT) Discover(key string) (*peer.PeerInfo, error) {
	log.DefaultLogger.Warn("=========== trying to resolve key " + key)
	ctx := context.Background()
	return r.GetPeerInfo(ctx, key)
}

// GetPeerInfo returns a peer's info from their id
func (r *DHT) GetPeerInfo(ctx context.Context, id string) (*peer.PeerInfo, error) {
	q := &query{
		dht:              r,
		id:               net.RandStringBytesMaskImprSrc(8),
		key:              id,
		queryType:        PeerInfoQuery,
		incomingPayloads: make(chan interface{}, 10),
		outgoingPayloads: make(chan interface{}, 10),
	}

	r.queries.Store(q.id, q)

	ctx, cf := context.WithTimeout(ctx, maxQueryTime)
	defer cf()

	go q.Run(ctx)

	for {
		select {
		case payload := <-q.outgoingPayloads:
			switch v := payload.(type) {
			case *peer.PeerInfo:
				return v, nil
			}
		case <-ctx.Done():
			return nil, ErrNotFound
		}
	}
}

// PutProviders adds a key of something we provide
// TODO Find a better name for this
func (r *DHT) PutProviders(ctx context.Context, key string) error {
	logger := log.Logger(ctx)
	provider := &Provider{
		ObjectIDs: []string{key},
	}
	so := provider.ToObject()
	if err := crypto.Sign(so, r.key); err != nil {
		return err
	}
	if err := r.store.PutProvider(provider); err != nil {
		return err
	}

	closestPeers, _ := r.FindPeersClosestTo(key, closestPeersToReturn)
	for _, closestPeer := range closestPeers {
		if err := r.exchange.Send(ctx, so, closestPeer.Address()); err != nil {
			logger.Debug("put providers could not send", zap.Error(err), zap.String("peerID", closestPeer.Fingerprint()))
		}
	}

	return nil
}

// GetProviders will look for peers that provide a key
func (r *DHT) GetProviders(ctx context.Context, key string) (chan *crypto.PublicKey, error) {
	q := &query{
		dht:              r,
		id:               net.RandStringBytesMaskImprSrc(8),
		key:              key,
		queryType:        ProviderQuery,
		incomingPayloads: make(chan interface{}, 10),
		outgoingPayloads: make(chan interface{}, 10),
	}

	r.queries.Store(q.id, q)

	go q.Run(ctx)

	out := make(chan *crypto.PublicKey, 1)
	go func(q *query, out chan *crypto.PublicKey) {
		defer close(out)
		for {
			select {
			case payload := <-q.outgoingPayloads:
				switch v := payload.(type) {
				case *Provider:
					// TODO do we need to check payload and id?
					if v.Signature != nil && v.Signature.PublicKey != nil {
						out <- v.Signature.PublicKey
					}
				}
			case <-time.After(maxQueryTime):
				return
			case <-ctx.Done():
				return
			}
		}
	}(q, out)

	return out, nil
}

func (r *DHT) GetAllProviders() (map[string][]string, error) {
	allProviders := map[string][]string{}
	providers, err := r.store.GetAllProviders()
	if err != nil {
		return nil, err
	}

	for _, provider := range providers {
		for _, objectID := range provider.ObjectIDs {
			if _, ok := allProviders[objectID]; !ok {
				allProviders[objectID] = []string{}
			}
			if provider.Signature == nil || provider.Signature.PublicKey == nil {
				continue
			}
			allProviders[objectID] = append(
				allProviders[objectID],
				provider.Signature.PublicKey.Fingerprint(),
			)
		}
	}
	return allProviders, nil
}
