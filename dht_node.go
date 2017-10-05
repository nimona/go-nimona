package dht

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	net "github.com/nimona/go-nimona-net"
)

const (
	protocolID = "dht-kad/v2"
	maxRetries = 3
)

type message struct {
	Recipient string `json:"-"`
	Type      string `json:"type"`
	Payload   []byte `json:"payload"`
	Retries   int    `json:"-"`
}

// DHTNode is the struct that implements the dht protocol
type DHTNode struct {
	localPeer net.Peer
	store     *Store
	net       net.Network
	messages  chan *message
	queries   map[string]*query
	streams   map[string]io.ReadWriteCloser
	lock      sync.RWMutex
}

func NewDHTNode(bps []net.Peer, lp net.Peer, nn net.Network) (*DHTNode, error) {
	// create new kv store
	st, _ := newStore()

	// Create DHT node
	nd := &DHTNode{
		localPeer: lp,
		store:     st,
		net:       nn,
		messages:  make(chan *message, 500),
		streams:   map[string]io.ReadWriteCloser{},
		queries:   map[string]*query{},
	}

	// register stream handler
	nd.net.RegisterStreamHandler(protocolID, nd.handleStream) // TODO(geoah) handle error?

	// Add bootstrap nodes
	for _, peer := range bps {
		if err := nd.storePeer(peer, true); err != nil {
			logrus.WithField("error", err).Error("new could not store peer")
		}
		if err := nd.putPeer(peer); err != nil {
			logrus.WithField("error", err).Error("new could not put peer")
		}
	}

	// TODO quit channel
	quit := make(chan struct{})

	// start refresh worker
	ticker := time.NewTicker(15 * time.Second)
	go func() {
		// refresh for the first time
		nd.refresh()
		// and then just wait
		for {
			select {
			case <-ticker.C:
				nd.refresh()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	// start messaging worker
	go func() {
		for msg := range nd.messages {
			st, err := nd.getStream(msg.Recipient)
			if err != nil {
				logrus.WithError(err).Warnf("Could not get stream for %s", msg.Recipient)
				if msg.Retries < maxRetries {
					msg.Retries++
					nd.messages <- msg
				}
				continue
			}
			bs, err := json.Marshal(msg)
			if err != nil {
				logrus.WithError(err).Warnf("Could not marshal message for %s, unrecoverable", msg.Recipient)
				continue
			}
			bs = append(bs, []byte("\n")...)
			logrus.WithField("recipient", msg.Recipient).WithField("line", string(bs)).Debugf("sending line")
			if _, err := st.Write(bs); err != nil {
				logrus.WithError(err).Warnf("Could not write to stream for %s", msg.Recipient)
				nd.lock.Lock()
				if _, ok := nd.streams[msg.Recipient]; ok {
					delete(nd.streams, msg.Recipient)
				}
				nd.lock.Unlock()
				if msg.Retries < maxRetries {
					msg.Retries++
					nd.messages <- msg
				}
				continue
			}
		}
	}()

	return nd, nil
}

func (nd *DHTNode) getStream(peerID string) (io.ReadWriteCloser, error) {
	nd.lock.Lock()
	defer nd.lock.Unlock()

	if stream, ok := nd.streams[peerID]; ok && stream != nil {
		// TODO Check if stream is still ok
		logrus.Debugf("Found stream for peer %s", peerID)
		return stream, nil
	}
	addr := peerID + "/" + protocolID
	logrus.WithField("addr", addr).Infof("Dialing peer for messabus.getStream")
	stream, err := nd.net.Dial(addr)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Created new stream for peer %s", peerID)
	nd.streams[peerID] = stream
	return stream, nil
}

func (nd *DHTNode) handleStream(protocolID string, stream io.ReadWriteCloser) error {
	sr := bufio.NewReader(stream)
	for {
		// read line
		line, err := sr.ReadString('\n')
		if err != nil {
			logrus.WithError(err).Errorf("Could not read")
			return err // TODO(geoah) Return?
		}
		logrus.WithField("line", line).Debugf("handleStream got line")

		// decode message
		msg := &message{}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			logrus.WithError(err).Warnf("Could not decode message")
			return err
		}

		// process message
		if err := nd.handleMessage(msg); err != nil {
			logrus.WithError(err).Warnf("Could not process message")
		}
	}
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

func (nd *DHTNode) handleMessage(msg *message) error {
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
		logrus.WithField("type", msg.Type).Info("Call type not implemented")
	}
	return nil
}

func (nd *DHTNode) Put(ctx context.Context, key, value string) error {
	logrus.Infof("Putting key %s", key)

	// create a put msg
	msgPut := &messagePut{
		OriginPeer: nd.GetLocalPeer(),
		Key:        key,
		Values:     []string{value},
	}
	// find nearest peers
	cps, err := nd.store.FindKeysNearestTo(KeyPrefixPeer, key, numPeersNear*10)
	if err != nil {
		logrus.WithError(err).Error("Put failed to find near peers")
		return err
	}
	for _, cp := range cps {
		// send message
		if err := nd.sendMessage(MessageTypePut, msgPut, trimKey(cp, KeyPrefixPeer)); err != nil {
			logrus.WithError(err).Warnf("Put could not send msg")
		}
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
	nd.lock.Lock()
	nd.queries[q.id] = q
	nd.lock.Unlock()

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

func (nd *DHTNode) sendMessage(msgType string, msg interface{}, peerID string) error {
	if peerID == nd.localPeer.ID {
		return nil
	}

	pl, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	nd.messages <- &message{
		Recipient: peerID,
		Payload:   pl,
		Type:      msgType,
	}

	return nil
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
		if err := nd.sendMessage(MessageTypePut, msgPut, msg.OriginPeer.ID); err != nil {
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
		if err := nd.sendMessage(MessageTypePut, msgPut, msg.OriginPeer.ID); err != nil {
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
