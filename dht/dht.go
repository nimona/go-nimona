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
	messengerExtention   = "dht"
	closestPeersToReturn = 8
	maxQueryTime         = time.Second * 5
)

// DHT is the struct that implements the dht protocol
type DHT struct {
	peerID         string
	store          *Store
	messenger      net.Messenger
	addressBook    net.PeerManager
	queries        sync.Map
	refreshBuckets bool
}

// NewDHT returns a new DHT from a messenger and peer manager
func NewDHT(messenger net.Messenger, pm net.PeerManager) (*DHT, error) {
	// create new kv store
	store, _ := newStore()

	// Create DHT node
	nd := &DHT{
		store:       store,
		messenger:   messenger,
		addressBook: pm,
		queries:     sync.Map{},
	}

	messenger.Handle("dht", nd.handleEnvelope)

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

		resp := EnvelopeGetPeerInfo{
			SenderPeerInfo: peerInfo.Envelope(),
			PeerID:         peerInfo.ID,
		}
		ctx := context.Background()
		peerIDs := getPeerIDsFromPeerInfos(closestPeers)
		envelope, err := net.NewEnvelope(PayloadTypeGetPeerInfo, peerIDs, resp)
		if err != nil {
			logrus.WithError(err).Warnf("refresh could not create envelope")
			return
		}
		if err := nd.messenger.Send(ctx, envelope); err != nil {
			logrus.WithError(err).Warnf("refresh could not send envelope")
			return
		}
		time.Sleep(time.Second * 30)
	}
}

func (nd *DHT) handleEnvelope(envelope *net.Envelope) error {
	// logrus.Debug("Got envelope", envelope.String())
	// if err := nd.addressBook.PutPeerInfoFromEnvelope(payload.SenderPeerInfo); err != nil {
	// 	logrus.WithError(err).Info("could not put sender peer info")
	// }
	contentType := envelope.Type
	switch contentType {
	case PayloadTypeGetPeerInfo:
		nd.handleGetPeerInfo(envelope)
	case PayloadTypePutPeerInfo:
		nd.handlePutPeerInfoFromEnvelope(envelope)
	case PayloadTypeGetProviders:
		nd.handleGetProviders(envelope)
	case PayloadTypePutProviders:
		nd.handlePutProviders(envelope)
	case PayloadTypeGetValue:
		nd.handleGetValue(envelope)
	case PayloadTypePutValue:
		nd.handlePutValue(envelope)
	default:
		logrus.WithField("envelope.PayloadType", contentType).Warn("Payload type not known")
		return nil
	}
	return nil
}

func (nd *DHT) handleGetPeerInfo(incEnvelope *net.Envelope) {
	payload, ok := incEnvelope.Payload.(EnvelopeGetPeerInfo)
	if !ok {
		logrus.Warn("expected EnvelopeGetPeerInfo, got ", reflect.TypeOf(incEnvelope.Payload))
		return
	}
	if err := nd.addressBook.PutPeerInfoFromEnvelope(payload.SenderPeerInfo); err != nil {
		logrus.WithError(err).Info("could not put sender peer info")
	}

	peerInfo, _ := nd.addressBook.GetPeerInfo(payload.PeerID)
	closestPeerInfos, _ := nd.FindPeersClosestTo(payload.PeerID, closestPeersToReturn)
	closestEnvelopes := getEnvelopesFromPeerInfos(closestPeerInfos)
	resp := EnvelopePutPeerInfoFromEnvelope{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().Envelope(),
		RequestID:      payload.RequestID,
		ClosestPeers:   closestEnvelopes,
	}
	if peerInfo != nil {
		resp.PeerInfo = peerInfo.Envelope
	}

	ctx := context.Background()
	to := []string{incEnvelope.Headers.Signer}
	envelope, err := net.NewEnvelope(PayloadTypePutPeerInfo, to, resp)
	if err != nil {
		logrus.WithError(err).Warnf("handleGetPeerInfo could not create envelope")
		return
	}
	if err := nd.messenger.Send(ctx, envelope); err != nil {
		logrus.WithError(err).Warnf("handleGetPeerInfo could not send envelope")
		return
	}
}

func (nd *DHT) handlePutPeerInfoFromEnvelope(incEnvelope *net.Envelope) {
	payload, ok := incEnvelope.Payload.(EnvelopePutPeerInfoFromEnvelope)
	if !ok {
		logrus.Warn("expected EnvelopePutPeerInfoFromEnvelope, got ", reflect.TypeOf(incEnvelope.Payload))
		return
	}
	if err := nd.addressBook.PutPeerInfoFromEnvelope(payload.SenderPeerInfo); err != nil {
		logrus.WithError(err).Info("could not put sender peer info")
	}

	for _, peerInfo := range payload.ClosestPeers {
		nd.addressBook.PutPeerInfoFromEnvelope(peerInfo)
	}

	if payload.RequestID == "" {
		return
	}

	q, exists := nd.queries.Load(payload.RequestID)
	if !exists {
		return
	}

	q.(*query).incomingEnvelopes <- payload
}

func (nd *DHT) handleGetProviders(incEnvelope *net.Envelope) {
	payload, ok := incEnvelope.Payload.(EnvelopeGetProviders)
	if !ok {
		logrus.Warn("expected EnvelopeGetProviders, got ", reflect.TypeOf(incEnvelope.Payload))
		return
	}
	if err := nd.addressBook.PutPeerInfoFromEnvelope(payload.SenderPeerInfo); err != nil {
		logrus.WithError(err).Info("could not put sender peer info")
	}

	providers, err := nd.store.GetProviders(payload.Key)
	if err != nil {
		return
	}

	closestPeerInfos, _ := nd.FindPeersClosestTo(payload.Key, closestPeersToReturn)
	closestEnvelopes := getEnvelopesFromPeerInfos(closestPeerInfos)
	resp := EnvelopePutProviders{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().Envelope(),
		RequestID:      payload.RequestID,
		Key:            payload.Key,
		PeerIDs:        providers,
		ClosestPeers:   closestEnvelopes,
	}

	ctx := context.Background()
	to := []string{payload.SenderPeerInfo.Headers.Signer}
	envelope, err := net.NewEnvelope(PayloadTypePutProviders, to, resp)
	if err != nil {
		logrus.WithError(err).Warnf("handleGetProviders could not create envelope")
		return
	}
	if err := nd.messenger.Send(ctx, envelope); err != nil {
		logrus.WithError(err).Warnf("handleGetProviders could not send envelope")
		return
	}
}

func (nd *DHT) handlePutProviders(incEnvelope *net.Envelope) {
	payload, ok := incEnvelope.Payload.(EnvelopePutProviders)
	if !ok {
		logrus.Warn("expected EnvelopePutProviders, got ", reflect.TypeOf(incEnvelope.Payload))
		return
	}
	if err := nd.addressBook.PutPeerInfoFromEnvelope(payload.SenderPeerInfo); err != nil {
		logrus.WithError(err).Info("could not put sender peer info")
	}

	for _, peerInfo := range payload.ClosestPeers {
		nd.addressBook.PutPeerInfoFromEnvelope(peerInfo)
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

	q.(*query).incomingEnvelopes <- payload
}

func (nd *DHT) handleGetValue(incEnvelope *net.Envelope) {
	payload, ok := incEnvelope.Payload.(EnvelopeGetValue)
	if !ok {
		logrus.Warn("expected EnvelopeGetValue, got ", reflect.TypeOf(incEnvelope.Payload))
		return
	}
	if err := nd.addressBook.PutPeerInfoFromEnvelope(payload.SenderPeerInfo); err != nil {
		logrus.WithError(err).Info("could not put sender peer info")
	}

	value, _ := nd.store.GetValue(payload.Key)

	closestPeerInfos, _ := nd.FindPeersClosestTo(payload.Key, closestPeersToReturn)
	closestEnvelopes := getEnvelopesFromPeerInfos(closestPeerInfos)
	resp := EnvelopePutValue{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().Envelope(),
		RequestID:      payload.RequestID,
		Key:            payload.Key,
		Value:          value,
		ClosestPeers:   closestEnvelopes,
	}

	ctx := context.Background()
	to := []string{payload.SenderPeerInfo.Headers.Signer}
	envelope, err := net.NewEnvelope(PayloadTypePutValue, to, resp)
	if err != nil {
		logrus.WithError(err).Warnf("handleGetValue could not create envelope")
		return
	}
	if err := nd.messenger.Send(ctx, envelope); err != nil {
		logrus.WithError(err).Warnf("handleGetValue could not send envelope")
		return
	}
}

func (nd *DHT) handlePutValue(incEnvelope *net.Envelope) {
	// TODO handle and log errors
	payload, ok := incEnvelope.Payload.(EnvelopePutValue)
	if !ok {
		logrus.Warn("expected EnvelopePutValue, got ", reflect.TypeOf(incEnvelope.Payload))
		return
	}
	if err := nd.addressBook.PutPeerInfoFromEnvelope(payload.SenderPeerInfo); err != nil {
		logrus.WithError(err).Info("could not put sender peer info")
	}

	for _, peerInfo := range payload.ClosestPeers {
		nd.addressBook.PutPeerInfoFromEnvelope(peerInfo)
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

	q.(*query).incomingEnvelopes <- payload
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
		dht:               nd,
		id:                net.RandStringBytesMaskImprSrc(8),
		key:               id,
		queryType:         PeerInfoQuery,
		incomingEnvelopes: make(chan interface{}),
		outgoingEnvelopes: make(chan interface{}),
	}

	nd.queries.Store(q.id, q)

	go q.Run(ctx)

	for {
		select {
		case value := <-q.outgoingEnvelopes:
			envelope := value.(*net.Envelope)
			nd.addressBook.PutPeerInfoFromEnvelope(envelope)
			return nd.addressBook.GetPeerInfo(envelope.Headers.Signer)
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
	resp := EnvelopePutValue{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().Envelope(),
		Key:            key,
		Value:          value,
	}

	closestPeerIDs := getPeerIDsFromPeerInfos(closestPeers)
	envelope, err := net.NewEnvelope(PayloadTypePutValue, closestPeerIDs, resp)
	if err != nil {
		logrus.WithError(err).Warnf("PutValue could not create envelope")
		return err
	}
	if err := nd.messenger.Send(ctx, envelope); err != nil {
		logrus.WithError(err).Warnf("PutValue could not send envelope")
		return err
	}

	return nil
}

func (nd *DHT) GetValue(ctx context.Context, key string) (string, error) {
	q := &query{
		dht:               nd,
		id:                net.RandStringBytesMaskImprSrc(8),
		key:               key,
		queryType:         ValueQuery,
		incomingEnvelopes: make(chan interface{}),
		outgoingEnvelopes: make(chan interface{}),
	}

	nd.queries.Store(q.id, q)

	go q.Run(ctx)

	for {
		select {
		case value := <-q.outgoingEnvelopes:
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
	resp := EnvelopePutProviders{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().Envelope(),
		Key:            key,
		PeerIDs:        []string{localPeerID},
	}

	closestPeerIDs := getPeerIDsFromPeerInfos(closestPeers)
	envelope, err := net.NewEnvelope(PayloadTypePutProviders, closestPeerIDs, resp)
	if err != nil {
		logrus.WithError(err).Warnf("PutProviders could not create envelope")
		return err
	}
	if err := nd.messenger.Send(ctx, envelope); err != nil {
		logrus.WithError(err).Warnf("PutProviders could not send envelope")
		return err
	}

	return nil
}

func (nd *DHT) GetProviders(ctx context.Context, key string) ([]string, error) {
	q := &query{
		dht:               nd,
		id:                net.RandStringBytesMaskImprSrc(8),
		key:               key,
		queryType:         ProviderQuery,
		incomingEnvelopes: make(chan interface{}),
		outgoingEnvelopes: make(chan interface{}),
	}

	nd.queries.Store(q.id, q)

	go q.Run(ctx)

	providers := []string{}
	for {
		select {
		case values := <-q.outgoingEnvelopes:
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

func getPeerIDsFromPeerInfos(peerInfos []*net.PeerInfo) []string {
	peerIDs := []string{}
	for _, peerInfo := range peerInfos {
		peerIDs = append(peerIDs, peerInfo.ID)
	}
	return peerIDs
}

func getEnvelopesFromPeerInfos(peerInfos []*net.PeerInfo) []*net.Envelope {
	envelopes := []*net.Envelope{}
	for _, peerInfo := range peerInfos {
		envelopes = append(envelopes, peerInfo.Envelope)
	}
	return envelopes
}
