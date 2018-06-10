package dht

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/nimona/go-nimona/peer"
	"github.com/nimona/go-nimona/wire"
)

var (
	ErrNotFound = errors.New("not found")
)

const (
	wireExtention        = "dht"
	closestPeersToReturn = 8
	maxQueryTime         = time.Second * 5
)

// DHT is the struct that implements the dht protocol
type DHT struct {
	peerID         string
	store          *Store
	wire           wire.Wire
	addressBook    peer.AddressBook
	queries        sync.Map
	refreshBuckets bool
}

func NewDHT(wr wire.Wire, pr peer.AddressBook) (*DHT, error) {
	// create new kv store
	store, _ := newStore()

	// Create DHT node
	nd := &DHT{
		store:       store,
		wire:        wr,
		addressBook: pr,
		queries:     sync.Map{},
	}

	wr.HandleExtensionEvents("dht", nd.handleMessage)

	go nd.refresh()

	return nd, nil
}

func (nd *DHT) refresh() {
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
			return
		}

		resp := messageGetPeerInfo{
			SenderPeerInfo: peerInfo.ToPeerInfo(),
			PeerID:         peerInfo.ID,
		}
		ctx := context.Background()
		peerIDs := getPeerIDsFromPeerInfos(closestPeers)
		nd.wire.Send(ctx, wireExtention, PayloadTypeGetPeerInfo, resp, peerIDs)
		time.Sleep(time.Second * 30)
	}
}

func (nd *DHT) handleMessage(message *wire.Message) error {
	// logrus.Debug("Got message", message.String())

	senderPeerInfo := &messageSenderPeerInfo{}
	if err := message.DecodePayload(senderPeerInfo); err == nil {
		nd.addressBook.PutPeerInfo(&senderPeerInfo.SenderPeerInfo)
	}

	switch message.PayloadType {
	case PayloadTypeGetPeerInfo:
		nd.handleGetPeerInfo(message)
	case PayloadTypePutPeerInfo:
		nd.handlePutPeerInfo(message)
	case PayloadTypeGetProviders:
		nd.handleGetProviders(message)
	case PayloadTypePutProviders:
		nd.handlePutProviders(message)
	case PayloadTypeGetValue:
		nd.handleGetValue(message)
	case PayloadTypePutValue:
		nd.handlePutValue(message)
	default:
		logrus.WithField("message.PayloadType", message.PayloadType).Warn("Payload type not known")
		return nil
	}
	return nil
}

func (nd *DHT) handleGetPeerInfo(message *wire.Message) {
	payload := &messageGetPeerInfo{}
	if err := message.DecodePayload(payload); err != nil {
		return
	}

	peerInfo, err := nd.addressBook.GetPeerInfo(payload.PeerID)
	if err != nil {
		return
	}

	closestPeers, _ := nd.FindPeersClosestTo(payload.PeerID, closestPeersToReturn)
	resp := messagePutPeerInfo{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().ToPeerInfo(),
		RequestID:      payload.RequestID,
		PeerID:         payload.PeerID,
		PeerInfo:       *peerInfo,
		ClosestPeers:   closestPeers,
	}

	ctx := context.Background()
	to := []string{payload.SenderPeerInfo.ID}
	nd.wire.Send(ctx, wireExtention, PayloadTypePutPeerInfo, resp, to)
}

func (nd *DHT) handlePutPeerInfo(message *wire.Message) {
	payload := &messagePutPeerInfo{}
	if err := message.DecodePayload(payload); err != nil {
		return
	}

	nd.addressBook.PutPeerInfo(&payload.PeerInfo)
	for _, peerInfo := range payload.ClosestPeers {
		nd.addressBook.PutPeerInfo(peerInfo)
	}

	if payload.RequestID == "" {
		return
	}

	q, exists := nd.queries.Load(payload.RequestID)
	if !exists {
		return
	}

	q.(*query).incomingMessages <- payload
}

func (nd *DHT) handleGetProviders(message *wire.Message) {
	payload := &messageGetProviders{}
	if err := message.DecodePayload(payload); err != nil {
		return
	}

	providers, err := nd.store.GetProviders(payload.Key)
	if err != nil {
		return
	}

	closestPeers, _ := nd.FindPeersClosestTo(payload.Key, closestPeersToReturn)
	resp := messagePutProviders{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().ToPeerInfo(),
		RequestID:      payload.RequestID,
		Key:            payload.Key,
		PeerIDs:        providers,
		ClosestPeers:   closestPeers,
	}

	ctx := context.Background()
	to := []string{payload.SenderPeerInfo.ID}
	nd.wire.Send(ctx, wireExtention, PayloadTypePutProviders, resp, to)
}

func (nd *DHT) handlePutProviders(message *wire.Message) {
	payload := &messagePutProviders{}
	if err := message.DecodePayload(payload); err != nil {
		return
	}

	for _, peerInfo := range payload.ClosestPeers {
		nd.addressBook.PutPeerInfo(peerInfo)
	}

	if err := nd.store.PutProvider(payload.Key, payload.PeerIDs...); err != nil {
		return
	}

	if payload.RequestID == "" {
		return
	}

	q, exists := nd.queries.Load(payload.RequestID)
	if !exists {
		return
	}

	q.(*query).incomingMessages <- payload
}

func (nd *DHT) handleGetValue(message *wire.Message) {
	payload := &messageGetValue{}
	if err := message.DecodePayload(payload); err != nil {
		return
	}

	value, _ := nd.store.GetValue(payload.Key)

	closestPeers, _ := nd.FindPeersClosestTo(payload.Key, closestPeersToReturn)
	resp := messagePutValue{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().ToPeerInfo(),
		RequestID:      payload.RequestID,
		Key:            payload.Key,
		Value:          value,
		ClosestPeers:   closestPeers,
	}

	ctx := context.Background()
	to := []string{payload.SenderPeerInfo.ID}
	nd.wire.Send(ctx, wireExtention, PayloadTypePutValue, resp, to)
}

func (nd *DHT) handlePutValue(message *wire.Message) {
	payload := &messagePutValue{}
	if err := message.DecodePayload(payload); err != nil {
		return
	}

	for _, peerInfo := range payload.ClosestPeers {
		nd.addressBook.PutPeerInfo(peerInfo)
	}

	if err := nd.store.PutValue(payload.Key, payload.Value); err != nil {
		return
	}

	if payload.RequestID == "" {
		return
	}

	q, exists := nd.queries.Load(payload.RequestID)
	if !exists {
		return
	}

	q.(*query).incomingMessages <- payload
}

// FindPeersClosestTo returns an array of n peers closest to the given key by xor distance
func (nd *DHT) FindPeersClosestTo(tk string, n int) ([]*peer.PeerInfo, error) {
	// place to hold the results
	rks := []*peer.PeerInfo{}

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

func (nd *DHT) GetPeerInfo(ctx context.Context, key string) (*peer.PeerInfo, error) {
	q := &query{
		dht:              nd,
		id:               wire.RandStringBytesMaskImprSrc(8),
		key:              key,
		queryType:        PeerInfoQuery,
		incomingMessages: make(chan interface{}),
		outgoingMessages: make(chan interface{}),
	}

	nd.queries.Store(q.id, q)

	go q.Run(ctx)

	for {
		select {
		case value := <-q.outgoingMessages:
			return value.(*peer.PeerInfo), nil
		case <-time.After(maxQueryTime):
			return nil, ErrNotFound
		case <-ctx.Done():
			return nil, ErrNotFound
		}
	}
}

func (nd *DHT) PutValue(ctx context.Context, key, value string) error {
	if err := nd.store.PutValue(key, value); err != nil {
		return err
	}

	closestPeers, _ := nd.FindPeersClosestTo(key, closestPeersToReturn)
	resp := messagePutValue{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().ToPeerInfo(),
		Key:            key,
		Value:          value,
	}

	closestPeerIDs := getPeerIDsFromPeerInfos(closestPeers)
	return nd.wire.Send(ctx, wireExtention, PayloadTypePutValue, resp, closestPeerIDs)
}

func (nd *DHT) GetValue(ctx context.Context, key string) (string, error) {
	q := &query{
		dht:              nd,
		id:               wire.RandStringBytesMaskImprSrc(8),
		key:              key,
		queryType:        ValueQuery,
		incomingMessages: make(chan interface{}),
		outgoingMessages: make(chan interface{}),
	}

	nd.queries.Store(q.id, q)

	go q.Run(ctx)

	for {
		select {
		case value := <-q.outgoingMessages:
			valueStr, ok := value.(string)
			if !ok {
				continue
			}
			return valueStr, nil
		case <-time.After(maxQueryTime):
			return "", ErrNotFound
		case <-ctx.Done():
			return "", ErrNotFound
		}
	}
}

// TODO Find a better name for this
func (nd *DHT) PutProviders(ctx context.Context, key string) error {
	localPeerID := nd.addressBook.GetLocalPeerInfo().ID
	if err := nd.store.PutProvider(key, localPeerID); err != nil {
		return err
	}

	closestPeers, _ := nd.FindPeersClosestTo(key, closestPeersToReturn)
	resp := messagePutProviders{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().ToPeerInfo(),
		Key:            key,
		PeerIDs:        []string{localPeerID},
	}

	closestPeerIDs := getPeerIDsFromPeerInfos(closestPeers)
	return nd.wire.Send(ctx, wireExtention, PayloadTypePutProviders, resp, closestPeerIDs)
}

func (nd *DHT) GetProviders(ctx context.Context, key string) ([]string, error) {
	q := &query{
		dht:              nd,
		id:               wire.RandStringBytesMaskImprSrc(8),
		key:              key,
		queryType:        ProviderQuery,
		incomingMessages: make(chan interface{}),
		outgoingMessages: make(chan interface{}),
	}

	nd.queries.Store(q.id, q)

	go q.Run(ctx)

	providers := []string{}
	for {
		select {
		case values := <-q.outgoingMessages:
			valuesStr, ok := values.([]string)
			if !ok {
				continue
			}
			providers = append(providers, valuesStr...)
		case <-time.After(maxQueryTime):
			return providers, nil
		case <-ctx.Done():
			return providers, nil
		}
	}
}

func (nd *DHT) GetAllProviders() (map[string][]string, error) {
	return nd.store.GetAllProviders()
}

func (nd *DHT) GetAllValues() (map[string]string, error) {
	return nd.store.GetAllValues()
}

func getPeerIDsFromPeerInfos(peerInfos []*peer.PeerInfo) []string {
	peerIDs := []string{}
	for _, peerInfo := range peerInfos {
		peerIDs = append(peerIDs, peerInfo.ID)
	}
	return peerIDs
}
