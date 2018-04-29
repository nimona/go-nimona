package dht

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/nimona/go-nimona/mesh"
	"github.com/nimona/go-nimona/wire"
)

const (
	wireExtention        = "dht"
	closestPeersToReturn = 8
)

// DHT is the struct that implements the dht protocol
type DHT struct {
	peerID         string
	store          *Store
	wire           *wire.Wire
	registry       mesh.Registry
	queries        sync.Map
	refreshBuckets bool
}

func NewDHT(wr *wire.Wire, pr mesh.Registry, peerID string, refreshBuckets bool, bootstrapAddresses ...string) (*DHT, error) {
	// create new kv store
	store, _ := newStore()

	// Create DHT node
	nd := &DHT{
		peerID:         peerID,
		store:          store,
		wire:           wr,
		registry:       pr,
		queries:        sync.Map{},
		refreshBuckets: refreshBuckets,
	}

	if err := nd.registry.PutPeerInfo(&mesh.PeerInfo{
		ID: "bootstrap",
		Protocols: map[string][]string{
			"wire": bootstrapAddresses,
		},
	}); err != nil {
		logrus.Warn("Could not put bootstrap peer", err)
	}

	wr.HandleExtensionEvents("dht", nd.handleMessage)

	go func() {
		time.Sleep(time.Second)
		for {
			nd.refresh()
			time.Sleep(time.Second * 15)
		}
	}()

	return nd, nil
}

func (nd *DHT) refresh() {
	peerInfo := nd.registry.GetLocalPeerInfo()
	cps, err := nd.FindPeersClosestTo(peerInfo.ID, closestPeersToReturn)
	if err != nil {
		logrus.WithError(err).Warnf("refresh could not get peers ids")
		return
	}

	resp := messagePutPeerInfo{
		PeerID:   peerInfo.ID,
		PeerInfo: *peerInfo,
	}
	ctx := context.Background()
	nd.wire.Send(ctx, wireExtention, PayloadTypePutPeerInfo, resp, cps)
}

func (nd *DHT) handleMessage(message *wire.Message) error {
	logrus.Info("Got message", message.String())
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

	peerInfo, err := nd.registry.GetPeerInfo(payload.PeerID)
	if err != nil {
		return
	}

	closestPeers, _ := nd.FindPeersClosestTo(payload.PeerID, closestPeersToReturn)
	resp := messagePutPeerInfo{
		RequestID:    payload.RequestID,
		PeerID:       payload.PeerID,
		PeerInfo:     *peerInfo,
		ClosestPeers: closestPeers,
	}

	ctx := context.Background()
	to := []string{message.From}
	nd.wire.Send(ctx, wireExtention, PayloadTypePutPeerInfo, resp, to)
}

func (nd *DHT) handlePutPeerInfo(message *wire.Message) {
	payload := &messagePutPeerInfo{}
	if err := message.DecodePayload(payload); err != nil {
		return
	}

	nd.registry.PutPeerInfo(&payload.PeerInfo)
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
		RequestID:    payload.RequestID,
		Key:          payload.Key,
		PeerIDs:      providers,
		ClosestPeers: closestPeers,
	}

	ctx := context.Background()
	to := []string{message.From}
	nd.wire.Send(ctx, wireExtention, PayloadTypePutProviders, resp, to)
}

func (nd *DHT) handlePutProviders(message *wire.Message) {
	payload := &messagePutProviders{}
	if err := message.DecodePayload(payload); err != nil {
		return
	}

	if err := nd.store.PutProvider(payload.Key, payload.PeerIDs...); err != nil {
		return
	}
}

func (nd *DHT) handleGetValue(message *wire.Message) {
	payload := &messageGetValue{}
	if err := message.DecodePayload(payload); err != nil {
		return
	}

	value, _ := nd.store.GetValue(payload.Key)

	closestPeers, _ := nd.FindPeersClosestTo(payload.Key, closestPeersToReturn)
	resp := messagePutValue{
		RequestID:    payload.RequestID,
		Key:          payload.Key,
		Value:        value,
		ClosestPeers: closestPeers,
	}

	ctx := context.Background()
	to := []string{message.From}
	nd.wire.Send(ctx, wireExtention, PayloadTypePutValue, resp, to)
}

func (nd *DHT) handlePutValue(message *wire.Message) {
	payload := &messagePutValue{}
	if err := message.DecodePayload(payload); err != nil {
		return
	}

	if err := nd.store.PutValue(payload.Key, payload.Value); err != nil {
		return
	}
}

// FindPeersClosestTo returns an array of n peers closest to the given key by xor distance
func (nd *DHT) FindPeersClosestTo(tk string, n int) ([]string, error) {
	// place to hold the results
	rks := []string{}

	htk := hash(tk)

	peerInfos, _ := nd.registry.GetAllPeerInfo()
	peerIDs := []string{}
	for _, peerInfo := range peerInfos {
		// remove self
		if nd.peerID == peerInfo.ID {
			continue
		}
		peerIDs = append(peerIDs, peerInfo.ID)
	}

	// slice to hold the distances
	dists := []distEntry{}
	for _, ik := range peerIDs {
		// calculate distance
		de := distEntry{
			key:  ik,
			dist: xor([]byte(htk), []byte(hash(ik))),
		}
		exists := false
		for _, ee := range dists {
			if ee.key == ik {
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
		rks = append(rks, de.key)
		n--
		if n == 0 {
			break
		}
	}

	return rks, nil
}

func (nd *DHT) Put(ctx context.Context, key, value string, labels map[string]string) error {
	if err := nd.store.PutValue(key, value); err != nil {
		return err
	}

	closestPeers, _ := nd.FindPeersClosestTo(key, closestPeersToReturn)
	resp := messagePutValue{
		Key:          key,
		Value:        value,
		ClosestPeers: closestPeers,
	}

	return nd.wire.Send(ctx, wireExtention, PayloadTypePutPeerInfo, resp, closestPeers)
}

func (nd *DHT) Get(ctx context.Context, key string) (string, error) {
	closestPeers, _ := nd.FindPeersClosestTo(key, closestPeersToReturn)
	req := messageGetValue{
		RequestID: mesh.RandStringBytesMaskImprSrc(8),
		Key:       key,
	}

	nd.wire.Send(ctx, wireExtention, PayloadTypeGetValue, req, closestPeers)
	return nd.store.GetValue(key)
}


func (nd *DHT) GetAllProviders() (map[string][]string, error) {
	return nd.store.GetAllProviders()
}

func (nd *DHT) GetAllValues() (map[string]string, error) {
	return nd.store.GetAllValues()
}
