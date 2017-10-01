package dht

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	messagebus "github.com/nimona/go-nimona-messagebus"
	net "github.com/nimona/go-nimona-net"
)

const (
	protocolID = "dht-kad/v1"
)

// DHTNode is the struct that implements the dht protocol
type DHTNode struct {
	localPeer  net.Peer
	store      *Store
	net        net.Network
	messageBus messagebus.MessageBus
	queries    map[string]*query
	mt         sync.RWMutex
}

func NewDHTNode(bps []net.Peer, lp net.Peer, nn net.Network) (*DHTNode, error) {
	// create new routing table
	st, _ := newStore()

	// create messagebud
	mb, err := messagebus.New(protocolID, nn)
	if err != nil {
		return nil, err
	}

	// Create DHT node
	nd := &DHTNode{
		localPeer:  lp,
		store:      st,
		net:        nn,
		messageBus: mb,
		queries:    map[string]*query{},
	}

	// Register message bus, message handler
	if err := mb.HandleMessage(nd.messageHandler); err != nil {
		return nil, err
	}

	// Add bootstrap nodes
	for _, peer := range bps {
		if err := nd.storePeer(peer, true); err != nil {
			logrus.WithField("error", err).Error("new could not store peer")
		}
		if err := nd.putPeer(peer); err != nil {
			logrus.WithField("error", err).Error("new could not put peer")
		}
	}

	// start refresh worker
	// ticker := time.NewTicker(15 * time.Second)
	// quit := make(chan struct{})
	go func() {
		// refresh for the first time
		nd.refresh()
		// and then just wait
		// for {
		// 	select {
		// 	case <-ticker.C:
		// 		nd.refresh()
		// 	case <-quit:
		// 		ticker.Stop()
		// 		return
		// 	}
		// }
	}()

	return nd, nil
}

func (nd *DHTNode) refresh() {
	logrus.Infof("Refreshing")
	cps, err := nd.store.FindKeysNearestTo(KeyPrefixPeer, nd.GetLocalPeer().ID, numPeersNear*10)
	if err != nil {
		logrus.WithError(err).Warnf("refresh could not get peers ids")
		return
	}
	ctx := context.Background()
	for _, cp := range cps {
		res, err := nd.Get(ctx, cp)
		if err != nil {
			logrus.WithError(err).WithField("peerID", cps).Warnf("refresh could not get for peer")
			continue
		}
		for range res {
			// just swallow channel results
		}
	}
}

func (nd *DHTNode) messageHandler(hash []byte, msg messagebus.Message) error {
	switch msg.Payload.Type {
	case MessageTypeGet:
		getMsg := &messageGet{}
		if err := json.Unmarshal(msg.Payload.Data, getMsg); err != nil {
			return err
		}
		nd.getHandler(getMsg)
	case MessageTypePut:
		putMsg := &messagePut{}
		if err := json.Unmarshal(msg.Payload.Data, putMsg); err != nil {
			return err
		}
		nd.putHandler(putMsg)
	default:
		logrus.Info("Call type not implemented")
	}
	return nil
}

func (nd *DHTNode) Get(ctx context.Context, key string) (chan string, error) {
	logrus.Infof("Searching for key %s", key)

	// create query
	// TODO query needs the context
	q := &query{
		id:               uuid.New().String(),
		dht:              nd,
		key:              key,
		contactedPeers:   []string{},
		results:          make(chan string, 100),
		incomingMessages: make(chan messagePut, 100),
		lock:             &sync.RWMutex{},
	}

	// and store it
	nd.mt.Lock()
	nd.queries[q.id] = q
	nd.mt.Unlock()

	// run query
	q.Run(ctx)

	// return results channel
	return q.results, nil
}

func (nd *DHTNode) GetPeer(ctx context.Context, id string) (net.Peer, error) {
	// get peer key
	res, err := nd.Get(ctx, getPeerKey(id))
	if err != nil {
		return net.Peer{}, err
	}

	// hold addresses
	addrs := []string{}

	// go through results and create addresses array
	for addr := range res {
		addrs = appendIfMissing(addrs, addr)
	}

	// check addrs
	if len(addrs) == 0 {
		return net.Peer{}, ErrPeerNotFound
	}

	return net.Peer{
		ID:        id,
		Addresses: addrs,
	}, nil
}

func (nd *DHTNode) sendMsgPeer(msgType string, msg interface{}, peerID string) error {
	if peerID == nd.localPeer.ID {
		return nil
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	pl := &messagebus.Payload{
		Creator: nd.localPeer.ID,
		Coded:   "json",
		Type:    msgType,
		Data:    data,
	}

	return nd.messageBus.Send(pl, []string{peerID})
}

func (nd *DHTNode) getHandler(msg *messageGet) {
	// origin peer is asking for a peer

	// store info on origin peer
	nd.storePeer(msg.OriginPeer, false)
	nd.putPeer(msg.OriginPeer)

	// check if we have the value of the key
	ks, err := nd.store.Get(msg.Key)
	if err != nil {
		logrus.WithField("msg", msg).Error("Failed to find nodes near")
		return
	}

	logrus.Infof("%+v", nd.store.pairs)

	// send them if we do
	if len(ks) > 0 {
		msgPut := &messagePut{
			QueryID:    msg.QueryID,
			OriginPeer: msg.OriginPeer,
			Key:        msg.Key,
			Values:     ks,
		}
		// send response
		if err := nd.sendMsgPeer(MessageTypePut, msgPut, msg.OriginPeer.ID); err != nil {
			logrus.WithError(err).Warnf("getHandler could not send msg")
		}
	}

	// find peers nearest peers that might have it
	cps, err := nd.store.FindKeysNearestTo(KeyPrefixPeer, msg.Key, numPeersNear)
	if err != nil {
		logrus.WithError(err).Error("getHandler could not find nearest peers")
		return
	}

	// give up if there are no peers
	if len(cps) == 0 {
		return
	}

	// send messages with closes peers
	for _, cp := range cps {
		cpid := trimKey(cp, KeyPrefixPeer)
		// skil us and original peer
		if cpid == msg.OriginPeer.ID || cpid == nd.GetLocalPeer().ID {
			continue
		}
		// get neighbor addresses
		addrs, err := nd.store.Get(cp)
		if err != nil {
			logrus.WithError(err).Warnf("getHandler could not get addrs")
			continue
		}
		// create a response
		msgPut := &messagePut{
			QueryID:    msg.QueryID,
			OriginPeer: msg.OriginPeer,
			Key:        cp,
			Values:     addrs,
		}
		// send response
		if err := nd.sendMsgPeer(MessageTypePut, msgPut, msg.OriginPeer.ID); err != nil {
			logrus.WithError(err).Warnf("getHandler could not send msg")
		}
	}
}

func (nd *DHTNode) putHandler(msg *messagePut) {
	// A peer we asked is informing us of a peer
	logrus.WithField("key", msg.Key).Infof("Got response")

	// check if this still a valid query
	if q, ok := nd.queries[msg.QueryID]; ok {
		q.incomingMessages <- *msg
	}

	// add values to our store
	if checkKey(msg.Key) {
		for _, v := range msg.Values {
			nd.store.Put(msg.Key, v, false)
		}
	}

	// check if this is a peer
	if strings.HasPrefix(msg.Key, KeyPrefixPeer) {
		pr, err := nd.gatherPeer(strings.Replace(msg.Key, KeyPrefixPeer, "", 1))
		if err != nil {
			logrus.WithError(err).Infof("putHandler could get pairs for putPeer")
			return
		}
		if err := nd.putPeer(pr); err != nil {
			logrus.WithError(err).Infof("putHandler could get putPeer")
			return
		}
	}
}

func (nd *DHTNode) gatherPeer(peerID string) (net.Peer, error) {
	addrs, err := nd.store.Get(peerID)
	if err != nil {
		return net.Peer{}, err
	}
	pr := net.Peer{
		ID:        peerID,
		Addresses: addrs,
	}
	return pr, nil
}

func (nd *DHTNode) putPeer(peer net.Peer) error {
	logrus.Infof("Adding peer to network peer=%v", peer)
	// add peer to network
	if err := nd.net.PutPeer(peer); err != nil {
		logrus.WithError(err).Warnf("Could not add peer to network")
		return err
	}
	logrus.Infof("PUT PEER %+v", peer)
	return nil
}

func (nd *DHTNode) storePeer(peer net.Peer, persistent bool) error {
	for _, addr := range peer.Addresses {
		logrus.WithField("k", getPeerKey(peer.ID)).WithField("v", addr).Infof("Adding peer addresses to kv")
		if err := nd.store.Put(getPeerKey(peer.ID), addr, persistent); err != nil {
			logrus.WithError(err).WithField("peerID", peer.ID).Warnf("storePeer could not put peer")
		}
	}
	return nil
}

func (nd *DHTNode) GetLocalPeer() net.Peer {
	return nd.localPeer
}
