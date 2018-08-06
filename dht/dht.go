package dht

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/nimona/go-nimona/net"
)

var (
	ErrNotFound = errors.New("not found")
)

const (
	exchangeExtention    = "dht"
	closestPeersToReturn = 8
	maxQueryTime         = time.Second * 5
)

// DHT is the struct that implements the dht protocol
type DHT struct {
	peerID         string
	store          *Store
	exchange       net.Exchange
	addressBook    net.AddressBooker
	queries        sync.Map
	refreshBuckets bool
}

// NewDHT returns a new DHT from a exchange and peer manager
func NewDHT(exchange net.Exchange, pm net.AddressBooker) (*DHT, error) {
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
		closestPeers, err := nd.FindPeersClosestTo(peerInfo.ID, closestPeersToReturn)
		if err != nil {
			logrus.WithError(err).Warnf("refresh could not get peers ids")
			time.Sleep(time.Second * 10)
			continue
		}

		// find peers to announce ourself to
		peerIDs := getPeerIDsFromPeerInfos(closestPeers)

		// announce our peer info to the closest peers
		if err := nd.exchange.Send(ctx, peerInfo.Block(), peerIDs...); err != nil {
			logrus.WithError(err).WithField("peer_ids", peerIDs).Warnf("refresh could not send block")
		}

		// HACK lookup our own peer info just so we can populate our peer table
		resp := PeerInfoRequest{
			PeerID: peerInfo.ID,
		}
		block := net.NewEphemeralBlock(PeerInfoRequestType, resp)
		if err := nd.exchange.Send(ctx, block, peerIDs...); err != nil {
			logrus.WithError(err).WithField("peer_ids", peerIDs).Warnf("refresh could not send block")
		}

		// sleep for a bit
		time.Sleep(time.Second * 30)
	}
}

func (nd *DHT) handleBlock(block *net.Block) error {
	contentType := block.Metadata.Type
	switch contentType {
	case PeerInfoRequestType:
		nd.handlePeerInfoRequest(block)
	case net.PeerInfoContentType:
		nd.handlePeerInfo(block)
	case ProviderRequestType:
		nd.handleProviderRequest(block)
	case ProviderType:
		nd.handleProvider(block)
	default:
		logrus.WithField("block.PayloadType", contentType).Warn("Payload type not known")
		return nil
	}
	return nil
}

func (nd *DHT) handlePeerInfoRequest(incBlock *net.Block) {
	ctx := context.Background()
	payload, ok := incBlock.Payload.(PeerInfoRequest)
	if !ok {
		logrus.Warn("expected PeerInfoRequest, got ", reflect.TypeOf(incBlock.Payload))
		return
	}

	peerInfo, _ := nd.addressBook.GetPeerInfo(payload.PeerID)
	if peerInfo != nil {
		nd.exchange.Send(ctx, peerInfo.Block, incBlock.Metadata.Signer)
		// TODO handle and log error
	}

	closestPeerInfos, _ := nd.FindPeersClosestTo(payload.PeerID, closestPeersToReturn)
	closestBlocks := getBlocksFromPeerInfos(closestPeerInfos)

	for _, block := range closestBlocks {
		if err := nd.exchange.Send(ctx, block, incBlock.Metadata.Signer); err != nil {
			logrus.WithError(err).Warnf("handlePeerInfoRequest could not send block")
			return
		}
	}
}

func (nd *DHT) handlePeerInfo(incBlock *net.Block) {
	payload, ok := incBlock.Payload.(net.PeerInfo)
	if !ok {
		logrus.Warn("expected PeerInfo, got ", reflect.TypeOf(incBlock.Payload))
		return
	}

	nd.addressBook.PutPeerInfoFromBlock(incBlock)

	rID := incBlock.GetHeader("requestID")
	if rID == "" {
		return
	}

	q, exists := nd.queries.Load(rID)
	if !exists {
		return
	}

	q.(*query).incomingPayloads <- payload
}

func (nd *DHT) handleProviderRequest(incBlock *net.Block) {
	ctx := context.Background()
	payload, ok := incBlock.Payload.(ProviderRequest)
	if !ok {
		logrus.Warn("expected ProviderRequest, got ", reflect.TypeOf(incBlock.Payload))
		return
	}

	providerBlocks, err := nd.store.GetProviders(payload.Key)
	if err != nil {
		// TODO handle and log error
		return
	}

	for _, providerBlock := range providerBlocks {
		nd.exchange.Send(ctx, providerBlock, incBlock.Metadata.Signer)
		// TODO handle and log error
	}

	closestPeerInfos, _ := nd.FindPeersClosestTo(payload.Key, closestPeersToReturn)
	closestBlocks := getBlocksFromPeerInfos(closestPeerInfos)

	for _, block := range closestBlocks {
		if err := nd.exchange.Send(ctx, block, incBlock.Metadata.Signer); err != nil {
			logrus.WithError(err).Warnf("handleProviderRequest could not send block")
			return
		}
	}
}

func (nd *DHT) handleProvider(incBlock *net.Block) {
	payload, ok := incBlock.Payload.(Provider)
	if !ok {
		logrus.Warn("expected Provider, got ", reflect.TypeOf(incBlock.Payload))
		return
	}

	if err := nd.store.PutProvider(incBlock); err != nil {
		// TODO log and handle error
	}

	rID := incBlock.GetHeader("requestID")
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
func (nd *DHT) FindPeersClosestTo(tk string, n int) ([]*net.PeerInfo, error) {
	// place to hold the results
	rks := []*net.PeerInfo{}

	htk := hash(tk)

	peerInfos, _ := nd.addressBook.GetAllPeerInfo()

	// slice to hold the distances
	dists := []distEntry{}
	for _, peerInfo := range peerInfos {
		// calculate distance
		de := distEntry{
			key:      peerInfo.ID,
			dist:     xor([]byte(htk), []byte(hash(peerInfo.ID))),
			peerInfo: peerInfo,
		}
		exists := false
		for _, ee := range dists {
			if ee.key == peerInfo.ID {
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
func (nd *DHT) GetPeerInfo(ctx context.Context, id string) (*net.PeerInfo, error) {
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
		case value := <-q.outgoingPayloads:
			block := value.(*net.Block)
			nd.addressBook.PutPeerInfoFromBlock(block)
			return nd.addressBook.GetPeerInfo(block.Metadata.Signer)
		case <-time.After(maxQueryTime):
			return nil, ErrNotFound
		case <-ctx.Done():
			return nil, ErrNotFound
		}
	}
}

// TODO Find a better name for this
func (nd *DHT) PutProviders(ctx context.Context, key string) error {
	providerBlock := net.NewEphemeralBlock(ProviderType, Provider{
		BlockIDs: []string{key},
	})

	signer := nd.addressBook.GetLocalPeerInfo()
	net.Sign(providerBlock, signer)

	if err := nd.store.PutProvider(providerBlock); err != nil {
		return err
	}

	block := net.NewEphemeralBlock(ProviderType, providerBlock)
	closestPeers, _ := nd.FindPeersClosestTo(key, closestPeersToReturn)
	closestPeerIDs := getPeerIDsFromPeerInfos(closestPeers)
	if err := nd.exchange.Send(ctx, block, closestPeerIDs...); err != nil {
		logrus.WithError(err).Warnf("PutProviders could not send block")
		return err
	}

	return nil
}

func (nd *DHT) GetProviders(ctx context.Context, key string) ([]string, error) {
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

	providers := []string{}
	for {
		select {
		case incBlock := <-q.outgoingPayloads:
			// TODO figure out why this might be nil
			if incBlock == nil {
				continue
			}
			block := incBlock.(*net.Block)
			// TODO do we need to check payload and id?
			// payload := block.Payload.(Provider)
			providers = append(providers, block.Metadata.Signer)
		case <-time.After(maxQueryTime):
			return providers, nil
		case <-ctx.Done():
			return providers, nil
		}
	}
}

func (nd *DHT) GetAllProviders() (map[string][]string, error) {
	providers := map[string][]string{}
	blocks, err := nd.store.GetAllProviders()
	if err != nil {
		return nil, err
	}

	for _, block := range blocks {
		payload := block.Payload.(Provider)
		for _, blockID := range payload.BlockIDs {
			if _, ok := providers[blockID]; !ok {
				providers[blockID] = []string{}
			}
			providers[blockID] = append(providers[blockID], block.Metadata.Signer)
		}
	}
	return providers, nil
}

func getPeerIDsFromPeerInfos(peerInfos []*net.PeerInfo) []string {
	peerIDs := []string{}
	for _, peerInfo := range peerInfos {
		peerIDs = append(peerIDs, peerInfo.ID)
	}
	return peerIDs
}

func getBlocksFromPeerInfos(peerInfos []*net.PeerInfo) []*net.Block {
	blocks := []*net.Block{}
	for _, peerInfo := range peerInfos {
		blocks = append(blocks, peerInfo.Block)
	}
	return blocks
}

func blocksOrNil(c []*net.Block) []*net.Block {
	if len(c) == 0 {
		return nil
	}

	return c
}
