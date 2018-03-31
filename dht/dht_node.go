package dht

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/nimona/go-nimona/mesh"
	"github.com/nimona/go-nimona/mutation"
	"github.com/nimona/go-nimona/net"
)

// DHT is the struct that implements the dht protocol
type DHT struct {
	peerID  string
	store   *Store
	queries sync.Map
	pubsub  mesh.PubSub
}

func NewDHT(ps mesh.PubSub, peerID string, bootstrapAddresses ...string) (*DHT, error) {
	// create new kv store
	store, _ := newStore()

	// Create DHT node
	nd := &DHT{
		peerID:  peerID,
		store:   store,
		pubsub:  ps,
		queries: sync.Map{},
	}

	for _, address := range bootstrapAddresses {
		if err := nd.putPeerAddress("bootstrap", "messaging", address, true); err != nil {
			return nil, err
		}
	}

	messages, _ := ps.Subscribe("messaging:dht:.*")
	go func() {
		for omsg := range messages {
			msg, ok := omsg.(mesh.Message)
			if !ok {
				continue
			}
			if msg.Sender == peerID {
				continue
			}
			if err := nd.handleMessage(msg); err != nil {
				fmt.Println("could not handle message", err)
			}
		}
	}()

	peerMessages, _ := ps.Subscribe("peer:.*")
	go func() {
		for omsg := range peerMessages {
			switch mut := omsg.(type) {
			case mutation.PeerProtocolDiscovered:
				cps, err := nd.store.FindPeersNearestTo(nd.peerID, numPeersNear)
				if err != nil {
					logrus.WithError(err).Warnf("bump could not get peers ids")
					continue
				}

				for _, cp := range cps {
					if err := nd.sendPutMessage(cp, mut.PeerID, mut.ProtocolAddress, map[string]string{
						"protocol": mut.ProtocolName,
					}); err != nil {
						fmt.Println("could not put subed addr")
					}
				}
			}
		}
	}()

	return nd, nil
}

// func (nd *DHT) refresh() {
// 	cps, err := nd.store.FindPeersNearestTo(nd.peerID, numPeersNear)
// 	if err != nil {
// 		logrus.WithError(err).Warnf("refresh could not get peers ids")
// 		return
// 	}

// 	localPeerAddresses, err := nd.store.Get(nd.peerID)
// 	if err != nil {
// 		return
// 	}

// 	logrus.Debugf("Refreshing with %v", cps)
// 	ctx := context.Background()
// 	for _, cp := range cps {
// 		for _, addr := range localPeerAddresses {
// 			if err := nd.sendPutMessage(cp, nd.peerID, addr.GetValue(), addr.GetLabels()); err != nil {
// 				fmt.Println("could not send own address on refresh", err)
// 			}
// 		}
// 		res, err := nd.Get(ctx, cp)
// 		if err != nil {
// 			logrus.WithError(err).WithField("peerID", cps).Warnf("refresh could not get for peer")
// 			continue
// 		}
// 		for range res {
// 			// just swallow channel results
// 		}
// 	}
// }

func (nd *DHT) handleMessage(msg mesh.Message) error {
	logrus.Info("Got message", msg.String())
	switch msg.Type {
	case MessageTypeGet:
		getMsg := &messageGet{}
		if err := json.Unmarshal(msg.Payload, getMsg); err != nil {
			return err
		}
		nd.getHandler(getMsg)
	case MessageTypePut:
		putMsg := &messagePut{}
		if err := json.Unmarshal(msg.Payload, putMsg); err != nil {
			return err
		}
		nd.putHandler(putMsg)
	default:
		logrus.WithField("type", msg.Type).Warn("Call type not implemented")
		return nil
	}
	return nil
}

func (nd *DHT) Put(ctx context.Context, key, value string, labels map[string]string) error {
	logrus.Debug("Putting key %s", key)

	// store this locally
	if err := nd.store.Put(key, value, labels, true); err != nil {
		logrus.WithError(err).Error("Put failed to store value locally")
	}

	// find nearest peers
	cps, err := nd.store.FindPeersNearestTo(key, numPeersNear)
	if err != nil {
		logrus.WithError(err).Error("Put failed to find near peers")
		return err
	}

	for _, cp := range cps {
		if err := nd.sendPutMessage(cp, key, value, labels); err != nil {
			return err
		}
	}

	return nil
}

func (nd *DHT) sendPutMessage(peerID, key, value string, labels map[string]string) error {
	// create a put msg
	msgPut := &messagePut{
		OriginPeerID: nd.peerID,
		Key:          key,
		Value:        value,
		Labels:       labels,
	}

	// send message
	if err := nd.sendMessage(MessageTypePut, msgPut, peerID); err != nil {
		logrus.WithError(err).Warnf("Put could not send msg")
	}
	// logrus.WithField("key", key).WithField("target", peerID).Debugf("Sent key to target")

	return nil
}

func (nd *DHT) Get(ctx context.Context, key string) (chan net.Record, error) {
	logrus.Debug("Searching for key %s", key)

	// create query
	// TODO query needs the context
	q := &query{
		id:               uuid.New().String(),
		dht:              nd,
		key:              key,
		labels:           map[string]string{},
		contactedPeers:   sync.Map{},
		results:          make(chan net.Record, 100),
		incomingMessages: make(chan messagePut, 100),
	}

	// and store it
	nd.queries.Store(q.id, q)

	// run query
	q.Run(ctx)

	// return results channel
	return q.results, nil
}

func (nd *DHT) Filter(ctx context.Context, key string, labels map[string]string) (chan net.Record, error) {
	logrus.Debug("Searching for key %s", key)

	// create query
	// TODO query needs the context
	q := &query{
		id:               uuid.New().String(),
		dht:              nd,
		key:              key,
		labels:           map[string]string{},
		contactedPeers:   sync.Map{},
		results:          make(chan net.Record, 100),
		incomingMessages: make(chan messagePut, 100),
	}

	// and store it
	nd.queries.Store(q.id, q)

	// run query
	q.Run(ctx)

	// return results channel
	return q.results, nil
}

// TODO(geoah) we might be better off accepting multiple peer ids to avoid re-marshaling every time
// TODO(geoah) don't get msg type as an arg, figure it out from the event type
func (nd *DHT) sendMessage(msgType string, event interface{}, peerID string) error {
	if peerID == nd.peerID {
		return nil
	}

	pl, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := mesh.Message{
		Recipient: peerID,
		Sender:    nd.peerID,
		Payload:   pl,
		Type:      msgType,
		Codec:     "json",
		Nonce:     mesh.RandStringBytesMaskImprSrc(8),
	}

	logrus.Info("Publishing message", msg.String())

	if err := nd.pubsub.Publish(msg, msgType); err != nil {
		return err
	}

	return nil
}

func (nd *DHT) getHandler(msg *messageGet) {
	// origin peer is asking for a value
	logger := logrus.
		WithField("origin.id", msg.OriginPeerID).
		WithField("key", msg.Key).
		WithField("query", msg.QueryID)
	logger.Debugf("Origin is asking for key")

	// check if we have the value of the key
	pairs, err := nd.store.Get(msg.Key)
	if err != nil {
		logger.Error("Failed to find nodes near")
		return
	}

	// send them if we do
	if len(pairs) > 0 {
		for _, pair := range pairs {
			msgPut := &messagePut{
				QueryID:      msg.QueryID,
				OriginPeerID: msg.OriginPeerID,
				Key:          msg.Key,
				Value:        pair.Value,
				Labels:       pair.Labels,
			}
			// send response
			if err := nd.sendMessage(MessageTypePut, msgPut, msg.OriginPeerID); err != nil {
				logger.WithError(err).Warnf("getHandler could not send msg")
			}
		}
		logger.Debugf("getHandler told origin about the values we knew")
	} else {
		logger.Debugf("getHandler does not know about this key")
	}

	// find peers nearest peers that might have it
	cps, err := nd.store.FindPeersNearestTo(msg.Key, numPeersNear)
	if err != nil {
		logger.WithError(err).Error("getHandler could not find nearest peers")
		return
	}

	logger.WithField("cps", cps).Debugf("Sending nearest peers")

	// give up if there are no peers
	if len(cps) == 0 {
		logger.Debugf("getHandler does not know any near peers")
		return
	}

	// send messages with closes peers
	for _, cp := range cps {
		// skip us and original peer
		if cp == msg.OriginPeerID {
			logger.Debugf("getHandler skipping origin")
			continue
		}
		if cp == nd.peerID {
			logger.Debugf("getHandler skipping local")
			continue
		}
		// get neighbor addresses
		addrs, err := nd.store.Get(cp)
		if err != nil {
			logger.WithError(err).Warnf("getHandler could not get addrs")
			continue
		}
		// create a response
		for _, addr := range addrs {
			msgPut := &messagePut{
				QueryID:      msg.QueryID,
				OriginPeerID: msg.OriginPeerID,
				Key:          cp,
				Value:        addr.Value,
				Labels: map[string]string{
					"protocol": "messaging",
				},
			}
			// send response
			if err := nd.sendMessage(MessageTypePut, msgPut, msg.OriginPeerID); err != nil {
				logger.WithError(err).Warnf("getHandler could not send msg")
			}
		}
	}
}

func (nd *DHT) putHandler(msg *messagePut) {
	// A peer we asked is informing us of a peer
	logger := logrus.
		WithField("key", msg.Key).
		WithField("query", msg.QueryID).
		WithField("origin", msg.OriginPeerID)
	logger.Debugf("Got response")

	// check if this still a valid query
	if q, ok := nd.queries.Load(msg.QueryID); ok {
		q.(*query).incomingMessages <- *msg
	}

	// TODO(geoah) lazy
	// add values to our store
	if msg.Labels["protocol"] != "" {
		nd.putPeerAddress(msg.Key, msg.Labels["protocol"], msg.Value, false)
	} else {
		nd.store.Put(msg.Key, msg.Value, msg.Labels, false)
	}
}

func (nd *DHT) putPeerAddress(peerID string, protocolName, protocolAddress string, pinned bool) error {
	if peerID == nd.peerID {
		return nil
	}

	labels := map[string]string{
		"protocol": protocolName,
	}

	logrus.Debugf("Adding peer to network id=%s protocol=%s address=%v", peerID, protocolName, protocolAddress)
	if err := nd.store.Put(peerID, protocolAddress, labels, true); err != nil {
		return err
	}

	nd.pubsub.Publish(mutation.PeerProtocolDiscovered{
		PeerID:          peerID,
		ProtocolName:    protocolName,
		ProtocolAddress: protocolAddress,
		Pinned:          pinned,
	}, mutation.PeerProtocolDiscoveredTopic)

	return nil
}

func (nd *DHT) GetLocalPairs() (map[string][]Pair, error) {
	return nd.store.GetAll()
}
