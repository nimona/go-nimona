package dht

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"

	"nimona.io/go/log"
	"nimona.io/go/net"
	"nimona.io/go/peers"
	"nimona.io/go/primitives"
)

var (
	ErrNotFound = errors.New("not found")
)

const (
	exchangeExtention    = "dht"
	closestPeersToReturn = 8
	maxQueryTime         = 2 * time.Second
)

// DHT is the struct that implements the dht protocol
type DHT struct {
	peerID         string
	store          *Store
	exchange       net.Exchange
	addressBook    *peers.AddressBook
	queries        sync.Map
	refreshBuckets bool
}

// NewDHT returns a new DHT from a exchange and peer manager
func NewDHT(exchange net.Exchange, pm *peers.AddressBook, addresses []string) (*DHT, error) {
	// create new kv store
	store, _ := newStore()

	// Create DHT node
	nd := &DHT{
		store:       store,
		exchange:    exchange,
		addressBook: pm,
		queries:     sync.Map{},
	}

	exchange.Handle("nimona.io/dht.**", nd.handleBlock)
	exchange.Handle("nimona.io/peer.info", nd.handleBlock)

	lk := pm.GetLocalPeerKey()
	for _, addr := range addresses {
		ctx := context.Background()
		req := &PeerInfoRequest{
			RequestID: net.RandStringBytesMaskImprSrc(8),
			PeerID:    lk.Thumbprint(),
		}
		reqBlock := req.Block()
		if err := exchange.Send(ctx, reqBlock, addr, primitives.SignWith(lk)); err != nil {
			log.Logger(ctx).Warn("could not send to bootstrap", zap.String("addr", addr), zap.Error(err))
		}
	}

	go nd.refresh()

	return nd, nil
}

func (nd *DHT) refresh() {
	ctx := context.Background()
	logger := log.Logger(ctx)
	// TODO this will be replaced when we introduce bucketing
	// TODO our init process is a bit messed up and addressBook doesn't know
	// about the peer's protocols instantly
	for {
		peerInfo := nd.addressBook.GetLocalPeerInfo()
		if len(peerInfo.Addresses) == 0 {
			time.Sleep(time.Second * 10)
			continue
		}

		closestPeers, err := nd.FindPeersClosestTo(peerInfo.Thumbprint(), closestPeersToReturn)
		if err != nil {
			logger.Warn("refresh could not get peers ids", zap.Error(err))
			time.Sleep(time.Second * 10)
			continue
		}

		// announce our peer info to the closest peers
		for _, closestPeer := range closestPeers {
			if err := nd.exchange.Send(ctx, peerInfo.Block(), closestPeer.Address()); err != nil {
				logger.Debug("refresh could not announce", zap.Error(err), zap.String("peerID", closestPeer.Thumbprint()))
			}
		}

		// HACK lookup our own peer info just so we can populate our peer table
		nd.GetPeerInfo(ctx, peerInfo.Thumbprint())

		// sleep for a bit
		time.Sleep(time.Second * 30)
	}
}

func (nd *DHT) handleBlock(block *primitives.Block) error {
	switch block.Type {
	case "nimona.io/dht.peer-info.request":
		v := &PeerInfoRequest{}
		v.FromBlock(block)
		nd.handlePeerInfoRequest(v)
	case "nimona.io/dht.peer-info.response":
		v := &PeerInfoResponse{}
		v.FromBlock(block)
		nd.handlePeerInfoResponse(v)
	case "nimona.io/dht.provider.request":
		v := &ProviderRequest{}
		v.FromBlock(block)
		nd.handleProviderRequest(v)
	case "nimona.io/dht.":
		v := &ProviderResponse{}
		v.FromBlock(block)
		nd.handleProviderResponse(v)
	case "nimona.io/peer.info":
		v := &peers.PeerInfo{}
		v.FromBlock(block)
		nd.handlePeerInfo(v)
	default:
		return nil
	}
	return nil
}

func (nd *DHT) handlePeerInfo(payload *peers.PeerInfo) {
	if err := nd.addressBook.PutPeerInfo(payload); err != nil {
		log.Logger(context.Background()).Error("could not handle peer info", zap.Error(err))
	}
}

func (nd *DHT) handlePeerInfoRequest(payload *PeerInfoRequest) {
	ctx := context.Background()
	logger := log.Logger(ctx)

	peerInfo, err := nd.addressBook.GetPeerInfo(payload.PeerID)
	if err != nil {
		// TODO handle and log error
	}

	closestPeerInfos, err := nd.FindPeersClosestTo(payload.PeerID, closestPeersToReturn)
	if err != nil {
		logger.Debug("could not get providers from local store", zap.Error(err))
		// TODO handle and log error
	}

	resp := &PeerInfoResponse{
		RequestID:    payload.RequestID,
		PeerInfo:     peerInfo,
		ClosestPeers: closestPeerInfos,
	}

	signer := nd.addressBook.GetLocalPeerKey()
	addr := "peer:" + payload.Signature.Key.Thumbprint()
	if err := nd.exchange.Send(ctx, resp.Block(), addr, primitives.SignWith(signer)); err != nil {
		logger.Debug("handleProviderRequest could not send block", zap.Error(err))
		return
	}
}

func (nd *DHT) handlePeerInfoResponse(payload *PeerInfoResponse) {
	ctx := context.Background()
	logger := log.Logger(ctx)
	for _, pi := range payload.ClosestPeers {
		if err := nd.addressBook.PutPeerInfo(pi); err != nil {
			logger.Error("could not handle closest peer from peerinfo response", zap.Error(err))
		}
	}

	if payload.PeerInfo != nil {
		if err := nd.addressBook.PutPeerInfo(payload.PeerInfo); err != nil {
			logger.Error("could not handle peer info from peerinfo response", zap.Error(err))
		}
	}

	rID := payload.RequestID
	if rID == "" {
		return
	}

	q, exists := nd.queries.Load(rID)
	if !exists {
		return
	}

	q.(*query).incomingPayloads <- payload.PeerInfo
}

func (nd *DHT) handleProviderRequest(payload *ProviderRequest) {
	ctx := context.Background()
	logger := log.Logger(ctx)

	providers, err := nd.store.GetProviders(payload.Key)
	if err != nil {
		logger.Debug("could not get providers from local store", zap.Error(err))
		// TODO handle and log error
	}

	closestPeerInfos, err := nd.FindPeersClosestTo(payload.Key, closestPeersToReturn)
	if err != nil {
		logger.Debug("could not get providers from local store", zap.Error(err))
		// TODO handle and log error
	}

	resp := &ProviderResponse{
		RequestID:    payload.RequestID,
		Providers:    providers,
		ClosestPeers: closestPeerInfos,
	}

	signer := nd.addressBook.GetLocalPeerKey()
	addr := "peer:" + payload.Signature.Key.Thumbprint()
	if err := nd.exchange.Send(ctx, resp.Block(), addr, primitives.SignWith(signer)); err != nil {
		logger.Warn("handleProviderRequest could not send block", zap.Error(err))
		return
	}
}

func (nd *DHT) handleProviderResponse(payload *ProviderResponse) {
	ctx := context.Background()
	logger := log.Logger(ctx)

	for _, provider := range payload.Providers {
		if err := nd.store.PutProvider(provider); err != nil {
			logger.Debug("could not store provider", zap.Error(err))
			// TODO handle error
		}
	}

	rID := payload.RequestID
	if rID == "" {
		return
	}

	q, exists := nd.queries.Load(rID)
	if !exists {
		return
	}

	q.(*query).incomingPayloads <- payload
}

// FindPeersClosestTo returns an array of n peers closest to the given key by xor distance
func (nd *DHT) FindPeersClosestTo(tk string, n int) ([]*peers.PeerInfo, error) {
	// place to hold the results
	rks := []*peers.PeerInfo{}

	htk := hash(tk)

	peerInfos, _ := nd.addressBook.GetAllPeerInfo()
	// slice to hold the distances
	dists := []distEntry{}
	for _, peerInfo := range peerInfos {
		peerInfoThumbprint := peerInfo.Thumbprint()
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

// GetPeerInfo returns a peer's info from their id
func (nd *DHT) GetPeerInfo(ctx context.Context, id string) (*peers.PeerInfo, error) {
	q := &query{
		dht:              nd,
		id:               net.RandStringBytesMaskImprSrc(8),
		key:              id,
		queryType:        PeerInfoQuery,
		incomingPayloads: make(chan interface{}, 10),
		outgoingPayloads: make(chan interface{}, 10),
	}

	nd.queries.Store(q.id, q)

	ctx, cf := context.WithTimeout(ctx, maxQueryTime)
	defer cf()

	go q.Run(ctx)

	for {
		select {
		case payload := <-q.outgoingPayloads:
			switch v := payload.(type) {
			case *peers.PeerInfo:
				return v, nil
			}
		case <-ctx.Done():
			return nil, ErrNotFound
		}
	}
}

// PutProviders adds a key of something we provide
// TODO Find a better name for this
func (nd *DHT) PutProviders(ctx context.Context, key string) error {
	logger := log.Logger(ctx)
	provider := &Provider{
		BlockIDs: []string{key},
	}
	signer := nd.addressBook.GetLocalPeerKey()
	providerBlock := provider.Block()
	if err := primitives.Sign(providerBlock, signer); err != nil {
		return err
	}
	provider.Signature = providerBlock.Signature
	if err := nd.store.PutProvider(provider); err != nil {
		return err
	}

	closestPeers, _ := nd.FindPeersClosestTo(key, closestPeersToReturn)
	for _, closestPeer := range closestPeers {
		if err := nd.exchange.Send(ctx, provider.Block(), closestPeer.Address(), primitives.SignWith(signer)); err != nil {
			logger.Debug("put providers could not send", zap.Error(err), zap.String("peerID", closestPeer.Thumbprint()))
		}
	}

	return nil
}

// GetProviders will look for peers that provide a key
func (nd *DHT) GetProviders(ctx context.Context, key string) (chan *primitives.Key, error) {
	q := &query{
		dht:              nd,
		id:               net.RandStringBytesMaskImprSrc(8),
		key:              key,
		queryType:        ProviderQuery,
		incomingPayloads: make(chan interface{}, 10),
		outgoingPayloads: make(chan interface{}, 10),
	}

	nd.queries.Store(q.id, q)

	go q.Run(ctx)

	out := make(chan *primitives.Key, 1)
	go func(q *query, out chan *primitives.Key) {
		defer close(out)
		for {
			select {
			case payload := <-q.outgoingPayloads:
				switch v := payload.(type) {
				case *Provider:
					// TODO do we need to check payload and id?
					// TODO Provider is not signed -- we don't seem to be storing the signature
					out <- v.Signature.Key
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

func (nd *DHT) GetAllProviders() (map[string][]string, error) {
	allProviders := map[string][]string{}
	providers, err := nd.store.GetAllProviders()
	if err != nil {
		return nil, err
	}

	for _, provider := range providers {
		for _, blockID := range provider.BlockIDs {
			if _, ok := allProviders[blockID]; !ok {
				allProviders[blockID] = []string{}
			}
			if provider.Signature == nil {
				continue
			}
			allProviders[blockID] = append(allProviders[blockID], provider.Signature.Key.Thumbprint())
		}
	}
	return allProviders, nil
}
