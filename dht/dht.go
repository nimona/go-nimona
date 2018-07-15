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

	messenger.Handle("dht", nd.handleMessage)

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

		resp := &MessageGetPeerInfo{
			SenderPeerInfo: peerInfo.Message(),
			PeerID:         peerInfo.ID,
		}
		ctx := context.Background()
		peerIDs := getPeerIDsFromPeerInfos(closestPeers)
		message, err := net.NewMessage(PayloadTypeGetPeerInfo, peerIDs, resp)
		if err != nil {
			logrus.WithError(err).Warnf("refresh could not create message")
			return
		}
		if err := nd.messenger.Send(ctx, message); err != nil {
			logrus.WithError(err).Warnf("refresh could not send message")
			return
		}
		time.Sleep(time.Second * 30)
	}
}

func (nd *DHT) handleMessage(message *net.Message) error {
	// logrus.Debug("Got message", message.String())
	// if err := nd.addressBook.PutPeerInfoFromMessage(payload.SenderPeerInfo); err != nil {
	// 	logrus.WithError(err).Info("could not put sender peer info")
	// }
	contentType := message.Headers.ContentType
	switch contentType {
	case PayloadTypeGetPeerInfo:
		nd.handleGetPeerInfo(message)
	case PayloadTypePutPeerInfo:
		nd.handlePutPeerInfoFromMessage(message)
	case PayloadTypeGetProviders:
		nd.handleGetProviders(message)
	case PayloadTypePutProviders:
		nd.handlePutProviders(message)
	case PayloadTypeGetValue:
		nd.handleGetValue(message)
	case PayloadTypePutValue:
		nd.handlePutValue(message)
	default:
		logrus.WithField("message.PayloadType", contentType).Warn("Payload type not known")
		return nil
	}
	return nil
}

func (nd *DHT) handleGetPeerInfo(incMessage *net.Message) {
	payload, ok := incMessage.Payload.(*MessageGetPeerInfo)
	if !ok {
		logrus.Warn("expected MessageGetPeerInfo, got ", reflect.TypeOf(incMessage.Payload))
		return
	}
	if err := nd.addressBook.PutPeerInfoFromMessage(payload.SenderPeerInfo); err != nil {
		logrus.WithError(err).Info("could not put sender peer info")
	}

	peerInfo, _ := nd.addressBook.GetPeerInfo(payload.PeerID)
	closestPeerInfos, _ := nd.FindPeersClosestTo(payload.PeerID, closestPeersToReturn)
	closestMessages := getMessagesFromPeerInfos(closestPeerInfos)
	resp := &MessagePutPeerInfoFromMessage{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().Message(),
		RequestID:      payload.RequestID,
		ClosestPeers:   closestMessages,
	}
	if peerInfo != nil {
		resp.PeerInfo = peerInfo.Message
	}

	ctx := context.Background()
	to := []string{incMessage.Headers.Signer}
	message, err := net.NewMessage(PayloadTypePutPeerInfo, to, resp)
	if err != nil {
		logrus.WithError(err).Warnf("handleGetPeerInfo could not create message")
		return
	}
	if err := nd.messenger.Send(ctx, message); err != nil {
		logrus.WithError(err).Warnf("handleGetPeerInfo could not send message")
		return
	}
}

func (nd *DHT) handlePutPeerInfoFromMessage(incMessage *net.Message) {
	payload, ok := incMessage.Payload.(*MessagePutPeerInfoFromMessage)
	if !ok {
		logrus.Warn("expected MessagePutPeerInfoFromMessage, got ", reflect.TypeOf(incMessage.Payload))
		return
	}
	if err := nd.addressBook.PutPeerInfoFromMessage(payload.SenderPeerInfo); err != nil {
		logrus.WithError(err).Info("could not put sender peer info")
	}

	for _, peerInfo := range payload.ClosestPeers {
		nd.addressBook.PutPeerInfoFromMessage(peerInfo)
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

func (nd *DHT) handleGetProviders(incMessage *net.Message) {
	payload, ok := incMessage.Payload.(*MessageGetProviders)
	if !ok {
		logrus.Warn("expected MessageGetProviders, got ", reflect.TypeOf(incMessage.Payload))
		return
	}
	if err := nd.addressBook.PutPeerInfoFromMessage(payload.SenderPeerInfo); err != nil {
		logrus.WithError(err).Info("could not put sender peer info")
	}

	providers, err := nd.store.GetProviders(payload.Key)
	if err != nil {
		return
	}

	closestPeerInfos, _ := nd.FindPeersClosestTo(payload.Key, closestPeersToReturn)
	closestMessages := getMessagesFromPeerInfos(closestPeerInfos)
	resp := &MessagePutProviders{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().Message(),
		RequestID:      payload.RequestID,
		Key:            payload.Key,
		PeerIDs:        providers,
		ClosestPeers:   closestMessages,
	}

	ctx := context.Background()
	to := []string{payload.SenderPeerInfo.Headers.Signer}
	message, err := net.NewMessage(PayloadTypePutProviders, to, resp)
	if err != nil {
		logrus.WithError(err).Warnf("handleGetProviders could not create message")
		return
	}
	if err := nd.messenger.Send(ctx, message); err != nil {
		logrus.WithError(err).Warnf("handleGetProviders could not send message")
		return
	}
}

func (nd *DHT) handlePutProviders(incMessage *net.Message) {
	payload, ok := incMessage.Payload.(*MessagePutProviders)
	if !ok {
		logrus.Warn("expected MessagePutProviders, got ", reflect.TypeOf(incMessage.Payload))
		return
	}
	if err := nd.addressBook.PutPeerInfoFromMessage(payload.SenderPeerInfo); err != nil {
		logrus.WithError(err).Info("could not put sender peer info")
	}

	for _, peerInfo := range payload.ClosestPeers {
		nd.addressBook.PutPeerInfoFromMessage(peerInfo)
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

func (nd *DHT) handleGetValue(incMessage *net.Message) {
	payload, ok := incMessage.Payload.(*MessageGetValue)
	if !ok {
		logrus.Warn("expected MessageGetValue, got ", reflect.TypeOf(incMessage.Payload))
		return
	}
	if err := nd.addressBook.PutPeerInfoFromMessage(payload.SenderPeerInfo); err != nil {
		logrus.WithError(err).Info("could not put sender peer info")
	}

	value, _ := nd.store.GetValue(payload.Key)

	closestPeerInfos, _ := nd.FindPeersClosestTo(payload.Key, closestPeersToReturn)
	closestMessages := getMessagesFromPeerInfos(closestPeerInfos)
	resp := &MessagePutValue{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().Message(),
		RequestID:      payload.RequestID,
		Key:            payload.Key,
		Value:          value,
		ClosestPeers:   closestMessages,
	}

	ctx := context.Background()
	to := []string{payload.SenderPeerInfo.Headers.Signer}
	message, err := net.NewMessage(PayloadTypePutValue, to, resp)
	if err != nil {
		logrus.WithError(err).Warnf("handleGetValue could not create message")
		return
	}
	if err := nd.messenger.Send(ctx, message); err != nil {
		logrus.WithError(err).Warnf("handleGetValue could not send message")
		return
	}
}

func (nd *DHT) handlePutValue(incMessage *net.Message) {
	// TODO handle and log errors
	payload, ok := incMessage.Payload.(*MessagePutValue)
	if !ok {
		logrus.Warn("expected MessagePutValue, got ", reflect.TypeOf(incMessage.Payload))
		return
	}
	if err := nd.addressBook.PutPeerInfoFromMessage(payload.SenderPeerInfo); err != nil {
		logrus.WithError(err).Info("could not put sender peer info")
	}

	for _, peerInfo := range payload.ClosestPeers {
		nd.addressBook.PutPeerInfoFromMessage(peerInfo)
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

func (nd *DHT) GetPeerInfo(ctx context.Context, key string) (*net.PeerInfo, error) {
	q := &query{
		dht:              nd,
		id:               net.RandStringBytesMaskImprSrc(8),
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
			return value.(*net.PeerInfo), nil
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
	resp := &MessagePutValue{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().Message(),
		Key:            key,
		Value:          value,
	}

	closestPeerIDs := getPeerIDsFromPeerInfos(closestPeers)
	message, err := net.NewMessage(PayloadTypePutValue, closestPeerIDs, resp)
	if err != nil {
		logrus.WithError(err).Warnf("PutValue could not create message")
		return err
	}
	if err := nd.messenger.Send(ctx, message); err != nil {
		logrus.WithError(err).Warnf("PutValue could not send message")
		return err
	}

	return nil
}

func (nd *DHT) GetValue(ctx context.Context, key string) (string, error) {
	q := &query{
		dht:              nd,
		id:               net.RandStringBytesMaskImprSrc(8),
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
	resp := &MessagePutProviders{
		SenderPeerInfo: nd.addressBook.GetLocalPeerInfo().Message(),
		Key:            key,
		PeerIDs:        []string{localPeerID},
	}

	closestPeerIDs := getPeerIDsFromPeerInfos(closestPeers)
	message, err := net.NewMessage(PayloadTypePutProviders, closestPeerIDs, resp)
	if err != nil {
		logrus.WithError(err).Warnf("PutProviders could not create message")
		return err
	}
	if err := nd.messenger.Send(ctx, message); err != nil {
		logrus.WithError(err).Warnf("PutProviders could not send message")
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

func getPeerIDsFromPeerInfos(peerInfos []*net.PeerInfo) []string {
	peerIDs := []string{}
	for _, peerInfo := range peerInfos {
		peerIDs = append(peerIDs, peerInfo.ID)
	}
	return peerIDs
}

func getMessagesFromPeerInfos(peerInfos []*net.PeerInfo) []*net.Message {
	messages := []*net.Message{}
	for _, peerInfo := range peerInfos {
		messages = append(messages, peerInfo.Message)
	}
	return messages
}
