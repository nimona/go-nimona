package dht

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/jinzhu/copier"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"

	"github.com/nimona/go-nimona/blocks"
	"github.com/nimona/go-nimona/log"
	"github.com/nimona/go-nimona/net"
	"github.com/nimona/go-nimona/peers"
)

var (
	ErrNotFound = errors.New("not found")
)

const (
	exchangeExtention    = "dht"
	closestPeersToReturn = 8
	maxQueryTime         = time.Second
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
func NewDHT(exchange net.Exchange, pm *peers.AddressBook) (*DHT, error) {
	// create new kv store
	store, _ := newStore()

	// Create DHT node
	nd := &DHT{
		store:       store,
		exchange:    exchange,
		addressBook: pm,
		queries:     sync.Map{},
	}

	exchange.Handle("dht", nd.handleBlock)
	exchange.Handle(peers.PeerInfoType, nd.handleBlock)

	go nd.refresh()

	return nd, nil
}

func (nd *DHT) refresh() {
	ctx := context.Background()
	// TODO this will be replaced when we introduce bucketing
	// TODO our init process is a bit messed up and addressBook doesn't know
	// about the peer's protocols instantly
	for len(nd.addressBook.GetLocalPeerInfo().Addresses) == 0 {
		time.Sleep(time.Millisecond * 250)
	}
	for {
		peerInfo := nd.addressBook.GetLocalPeerInfo()
		closestPeers, err := nd.FindPeersClosestTo(peerInfo.Thumbprint(), closestPeersToReturn)
		if err != nil {
			logrus.WithError(err).Warnf("refresh could not get peers ids")
			time.Sleep(time.Second * 10)
			continue
		}

		signer := nd.addressBook.GetLocalPeerInfo().Key

		// announce our peer info to the closest peers
		for _, closestPeer := range closestPeers {
			if err := nd.exchange.Send(ctx, closestPeer, peerInfo.Key, blocks.SignWith(signer)); err != nil {
				panic(err)
			}
		}

		// HACK lookup our own peer info just so we can populate our peer table
		nd.GetPeerInfo(ctx, peerInfo.Thumbprint())

		// sleep for a bit
		time.Sleep(time.Second * 30)
	}
}

func (nd *DHT) handleBlock(payload interface{}) error {
	switch v := payload.(type) {
	case *PeerInfoRequest:
		nd.handlePeerInfoRequest(v)
	case *peers.PeerInfo:
		nd.handlePeerInfo(v)
	case *ProviderRequest:
		nd.handleProviderRequest(v)
	case *Provider:
		nd.handleProvider(v)
	default:
		return nil
	}
	return nil
}

func (nd *DHT) handlePeerInfoRequest(payload *PeerInfoRequest) {
	ctx := context.Background()
	signer := nd.addressBook.GetLocalPeerInfo().Key
	peerInfo, _ := nd.addressBook.GetPeerInfo(payload.PeerID)
	if peerInfo != nil {
		nd.exchange.Send(ctx, peerInfo, payload.Signature.Key, blocks.SignWith(signer))
		// TODO handle and log error
	}

	closestPeerInfos, _ := nd.FindPeersClosestTo(payload.PeerID, closestPeersToReturn)
	for _, closestPeerInfo := range closestPeerInfos {
		if err := nd.exchange.Send(ctx, closestPeerInfo, payload.Signature.Key, blocks.SignWith(signer)); err != nil {
			logrus.WithError(err).Warnf("handlePeerInfoRequest could not send block")
			return
		}
	}
}

func (nd *DHT) handlePeerInfo(payload *peers.PeerInfo) {
	// TODO handle error
	nd.addressBook.PutPeerInfo(payload)

	// TODO headers
	// rID := incBlock.GetHeader("requestID")
	// if rID == "" {
	// 	return
	// }

	// q, exists := nd.queries.Load(rID)
	// if !exists {
	// 	return
	// }

	// q.(*query).incomingPayloads <- incBlock
}

func (nd *DHT) handleProviderRequest(payload *ProviderRequest) {
	ctx := context.Background()
	logger := log.Logger(ctx)

	signer := nd.addressBook.GetLocalPeerInfo().Key
	providers, err := nd.store.GetProviders(payload.Key)
	if err != nil {
		logger.Debug("could not get providers from local store", zap.Error(err))
		// TODO handle and log error
		return
	}

	for _, provider := range providers {
		copy := Provider{}
		copier.Copy(&copy, &provider)
		// cProviderBlock.SetHeader("requestID", payload.RequestID)
		logger.Debug("found provider block")
		nd.exchange.Send(ctx, copy, payload.Signature.Key, blocks.SignWith(signer))
		// TODO handle and log error
	}

	closestPeerInfos, _ := nd.FindPeersClosestTo(payload.Key, closestPeersToReturn)
	for _, peerInfo := range closestPeerInfos {
		copy := peers.PeerInfo{}
		copier.Copy(&copy, &peerInfo)
		// cBlock.SetHeader("requestID", payload.RequestID)
		// logger.Debug("sending provider block", zap.String("blockID", blocks.BestEffortID(cBlock)))
		if err := nd.exchange.Send(ctx, copy, payload.Signature.Key, blocks.SignWith(signer)); err != nil {
			logger.Warn("handleProviderRequest could not send block", zap.Error(err))
			return
		}
	}
}

func (nd *DHT) handleProvider(payload *Provider) {
	ctx := context.Background()
	logger := log.Logger(ctx)

	// logger.Debug("handling provider",
	// zap.String("blockID", blocks.BestEffortID(incBlock)),
	// zap.String("requestID", incBlock.GetHeader("requestID")))

	if err := nd.store.PutProvider(payload); err != nil {
		logger.Debug("could not store provider", zap.Error(err))
		// TODO handle error
	}

	// rID := incBlock.GetHeader("requestID")
	// if rID == "" {
	// 	return
	// }

	// q, exists := nd.queries.Load(rID)
	// if !exists {
	// 	return
	// }

	// q.(*query).incomingPayloads <- incBlock
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
		incomingPayloads: make(chan interface{}),
		outgoingPayloads: make(chan interface{}),
	}

	nd.queries.Store(q.id, q)

	go q.Run(ctx)

	for {
		select {
		case payload := <-q.outgoingPayloads:
			// TODO handle error
			switch v := payload.(type) {
			case *peers.PeerInfo:
				nd.addressBook.PutPeerInfo(v)
				return nd.addressBook.GetPeerInfo(v.Thumbprint())
			}
		case <-time.After(maxQueryTime):
			return nil, ErrNotFound
		case <-ctx.Done():
			return nil, ErrNotFound
		}
	}
}

// PutProviders adds a key of something we provide
// TODO Find a better name for this
func (nd *DHT) PutProviders(ctx context.Context, key string) error {
	provider := &Provider{
		BlockIDs: []string{key},
	}
	signer := nd.addressBook.GetLocalPeerInfo().Key
	if err := nd.store.PutProvider(provider); err != nil {
		return err
	}

	closestPeers, _ := nd.FindPeersClosestTo(key, closestPeersToReturn)
	for _, closestPeer := range closestPeers {
		if err := nd.exchange.Send(ctx, provider, closestPeer.Signature.Key, blocks.SignWith(signer)); err != nil {
			panic(err)
		}
	}

	return nil
}

// GetProviders will look for peers that provide a key
func (nd *DHT) GetProviders(ctx context.Context, key string) (chan string, error) {
	q := &query{
		dht:              nd,
		id:               net.RandStringBytesMaskImprSrc(8),
		key:              key,
		queryType:        ProviderQuery,
		incomingPayloads: make(chan interface{}),
		outgoingPayloads: make(chan interface{}),
	}

	nd.queries.Store(q.id, q)

	go q.Run(ctx)

	out := make(chan string, 1)
	go func(q *query, out chan string) {
		defer close(out)
		for {
			select {
			case payload := <-q.outgoingPayloads:
				switch v := payload.(type) {
				case Provider:
					// TODO do we need to check payload and id?
					out <- v.Signature.Key.Thumbprint()
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

// func getPeerIDsFromPeerInfos(peerInfos []*peers.PeerInfo) []string {
// 	peerIDs := []string{}
// 	for _, peerInfo := range peerInfos {
// 		peerIDs = append(peerIDs, peerInfo.Thumbprint())
// 	}
// 	return peerIDs
// }

// func getBlocksFromPeerInfos(peerInfos []*peers.PeerInfo) []*blocks.Block {
// 	blocks := []*blocks.Block{}
// 	for _, peerInfo := range peerInfos {
// 		blocks = append(blocks, peerInfo.Block)
// 	}
// 	return blocks
// }

func blocksOrNil(c []*blocks.Block) []*blocks.Block {
	if len(c) == 0 {
		return nil
	}

	return c
}
