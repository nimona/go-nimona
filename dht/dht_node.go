package dht

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/nimona/go-nimona-fabric"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	net "github.com/nimona/go-nimona-net"
)

const (
	protocolID = "dht-kad/v2"
)

type message struct {
	Recipient string `json:"-"`
	Type      string `json:"type"`
	Payload   []byte `json:"payload"`
	Retries   int    `json:"-"`
}

// DHT is the struct that implements the dht protocol
type DHT struct {
	localPeer     *messagePeer
	store         *Store
	messages      chan *message
	queries       map[string]*query
	peerAddresses map[string][]string
	lock          sync.RWMutex
	net           *fabric.Fabric
}

func NewDHT(bps map[string][]string, localPeerID string, nn *fabric.Fabric) (*DHT, error) {
	// create new kv store
	st, _ := newStore()

	// Create DHT node
	nd := &DHT{
		localPeer: &messagePeer{
			ID:        localPeerID,
			Addresses: nn.GetAddresses(),
		},
		store:         st,
		net:           nn,
		messages:      make(chan *message, 500),
		peerAddresses: bps,
		queries:       map[string]*query{},
	}

	// register stream handler
	// nd.net.RegisterStreamHandler(protocolID, nd.handleStream) // TODO(geoah) handle error?

	// Add bootstrap nodes
	for peerID, addresses := range bps {
		// if err := nd.storePeer(peer, true); err != nil {
		// 	logrus.WithField("error", err).Error("new could not store peer")
		// }
		if err := nd.putPeer(peerID, addresses); err != nil {
			logrus.WithField("error", err).Error("new could not put peer")
		}
	}

	// TODO quit channel
	// quit := make(chan struct{})

	// start refresh worker
	// ticker := time.NewTicker(30 * time.Second)
	// go func() {
	// 	// refresh for the first time
	// 	nd.refresh()
	// 	// and then just wait
	// 	for {
	// 		select {
	// 		case <-ticker.C:
	// 			nd.refresh()
	// 		case <-quit:
	// 			ticker.Stop()
	// 			return
	// 		}
	// 	}
	// }()

	// start messaging worker
	// TODO move this to a separate method
	go func() {
		for msg := range nd.messages {
			logger := logrus.
				WithField("target", msg.Recipient).
				WithField("type", msg.Type).
				WithField("target", msg.Recipient)

			logger.Debugf("Trying to send message to peer")
			st, err := nd.getStream(msg.Recipient)
			if err != nil {
				logger.WithError(err).Warnf("Could not get stream")
				continue
			}

			logger.Debugf("- Got stream")
			bs, err := json.Marshal(msg)
			if err != nil {
				logger.WithError(err).Warnf("Could not marshal message")
				continue
			}

			bs = append(bs, []byte("\n")...)
			logger.WithField("line", string(bs)).Debugf("sending line")
			if _, err := st.Write(bs); err != nil {
				logger.WithError(err).Warnf("Could not write to stream")
				// if _, ok := nd.streams[msg.Recipient]; ok {
				// 	delete(nd.streams, msg.Recipient)
				// }
			}
			logrus.Infof("Message sent")
		}
	}()

	return nd, nil
}

func (nd *DHT) getStream(peerID string) (io.ReadWriteCloser, error) {
	// TODO re-introduce caching
	ctx := context.Background()
	_, c, err := nd.net.DialContext(ctx, nd.peerAddresses[peerID][0])
	return c, err
	// if stream, ok := nd.streams[peerID]; ok && stream != nil {
	// 	// TODO Check if stream is still ok
	// 	logrus.Debugf("Found stream for peer %s", peerID)
	// 	return stream, nil
	// }
	// addr := peerID + "/" + protocolID
	// logrus.WithField("addr", addr).Infof("Dialing peer for messabus.getStream")
	// stream, err := nd.net.Dial(addr)
	// if err != nil {
	// 	return nil, err
	// }
	// logrus.Debugf("Created new stream for peer %s", peerID)
	// nd.streams[peerID] = stream
	// return stream, nil
}

func (nd *DHT) handleStream(protocolID string, stream io.ReadWriteCloser) error {
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

func (nd *DHT) refresh() {
	logrus.Infof("Refreshing")
	cps, err := nd.store.FindKeysNearestTo(KeyPrefixPeer, nd.localPeer.ID, numPeersNear)
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
	pairs, err := nd.store.GetAll()
	if err != nil {
		logrus.WithError(err).Warnf("refresh could not get all pairs")
		return
	}

	for _, prs := range pairs {
		for _, pr := range prs {
			if pr.Persistent {
				nd.sendPutMessage(pr.Key, pr.Value)
			}
		}
	}
}

func (nd *DHT) handleMessage(msg *message) error {
	logrus.WithField("payload", string(msg.Payload)).WithField("type", msg.Type).Infof("Trying to handle message")
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

func (nd *DHT) Put(ctx context.Context, key, value string) error {
	logrus.Infof("Putting key %s", key)

	// store this locally
	if err := nd.store.Put(key, value, true); err != nil {
		logrus.WithError(err).Error("Put failed to store value locally")
	}

	return nd.sendPutMessage(key, value)
}

func (nd *DHT) sendPutMessage(key, value string) error {
	// get fresh local peer addresses
	localPeer := &messagePeer{
		ID:        nd.localPeer.ID,
		Addresses: nd.net.GetAddresses(),
	}
	// create a put msg
	msgPut := &messagePut{
		OriginPeer: localPeer,
		Key:        key,
		Values:     []string{value},
	}

	// find nearest peers
	cps, err := nd.store.FindKeysNearestTo(KeyPrefixPeer, key, numPeersNear*10)
	if err != nil {
		logrus.WithError(err).Error("Put failed to find near peers")
		return err
	}
	fmt.Println("!!!", cps)
	for _, cp := range cps {
		// send message
		if err := nd.sendMessage(MessageTypePut, msgPut, trimKey(cp, KeyPrefixPeer)); err != nil {
			logrus.WithError(err).Warnf("Put could not send msg")
			continue
		}
		logrus.WithField("key", key).WithField("target", cp).Infof("Sent key to target")
	}

	return nil
}

func (nd *DHT) Get(ctx context.Context, key string) (chan string, error) {
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

	fmt.Println("------ query", q.id)

	// and store it
	nd.lock.Lock()
	nd.queries[q.id] = q
	nd.lock.Unlock()

	// run query
	q.Run(ctx)

	// return results channel
	return q.results, nil
}

func (nd *DHT) GetPeer(ctx context.Context, id string) ([]string, error) {
	// get peer key
	res, err := nd.Get(ctx, getPeerKey(id))
	if err != nil {
		return nil, err
	}

	// hold addresses
	addrs := []string{}

	// go through results and create addresses array
	for addr := range res {
		addrs = appendIfMissing(addrs, addr)
	}

	// check addrs
	if len(addrs) == 0 {
		return nil, ErrPeerNotFound
	}

	return addrs, nil
}

func (nd *DHT) sendMessage(msgType string, payload interface{}, peerID string) error {
	if peerID == nd.localPeer.ID {
		return nil
	}

	pl, err := json.Marshal(payload)
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

func (nd *DHT) getHandler(msg *messageGet) {
	// origin peer is asking for a value
	logger := logrus.
		WithField("origin", msg.OriginPeer.ID).
		WithField("key", msg.Key).
		WithField("query", msg.QueryID)
	logger.Infof("Origin is asking for key")

	// store info on origin peer
	// nd.storePeer(msg.OriginPeer, false)
	nd.putPeer(msg.OriginPeer.ID, msg.OriginPeer.Addresses)

	// check if we have the value of the key
	ks, err := nd.store.Get(msg.Key)
	if err != nil {
		logger.Error("Failed to find nodes near")
		return
	}

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
			logger.WithError(err).Warnf("getHandler could not send msg")
		}
		logger.Infof("getHandler told origin about the value")
	} else {
		logger.Infof("getHandler does not know about this key")
	}

	// find peers nearest peers that might have it
	cps, err := nd.store.FindKeysNearestTo(KeyPrefixPeer, msg.Key, numPeersNear)
	if err != nil {
		logger.WithError(err).Error("getHandler could not find nearest peers")
		return
	}

	logger.WithField("cps", cps).Infof("Sending nearest peers")

	// give up if there are no peers
	if len(cps) == 0 {
		logger.Infof("getHandler does not know any near peers")
		return
	}

	// send messages with closes peers
	for _, cp := range cps {
		cpid := trimKey(cp, KeyPrefixPeer)
		// skip us and original peer
		if cpid == msg.OriginPeer.ID {
			logger.Debugf("getHandler skipping origin")
			continue
		}
		// if cpid == nd.GetLocalPeer().ID {
		if cpid == nd.localPeer.ID {
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
		msgPut := &messagePut{
			QueryID:    msg.QueryID,
			OriginPeer: msg.OriginPeer,
			Key:        cp,
			Values:     addrs,
		}
		// send response
		if err := nd.sendMessage(MessageTypePut, msgPut, msg.OriginPeer.ID); err != nil {
			logger.WithError(err).Warnf("getHandler could not send msg")
		}
	}
}

func (nd *DHT) putHandler(msg *messagePut) {
	// A peer we asked is informing us of a peer
	logger := logrus.
		WithField("key", msg.Key).
		WithField("query", msg.QueryID).
		WithField("origin", msg.OriginPeer.ID)
	logger.Infof("Got response")

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
		peerID := strings.Replace(msg.Key, KeyPrefixPeer, "", 1)
		// pr, err := nd.gatherPeer(strings.Replace(msg.Key, KeyPrefixPeer, "", 1))
		// if err != nil {
		// 	logger.WithError(err).Infof("putHandler could get pairs for putPeer")
		// 	return
		// }
		if err := nd.putPeer(peerID, msg.Values); err != nil {
			logger.WithError(err).Infof("putHandler could putPeer")
			return
		}
	}

	if err := nd.putPeer(msg.OriginPeer.ID, msg.OriginPeer.Addresses); err != nil {
		logger.WithError(err).Infof("putHandler could putPeer for origin")
		return
	}
}

// func (nd *DHT) gatherPeer(peerID string) (net.Peer, error) {
// 	addrs, err := nd.store.Get(getPeerKey(peerID))
// 	if err != nil {
// 		return net.Peer{}, err
// 	}
// 	pr := net.Peer{
// 		ID:        peerID,
// 		Addresses: addrs,
// 	}
// 	return pr, nil
// }

func (nd *DHT) putPeer(peerID string, peerAddresses []string) error {
	if peerID == nd.localPeer.ID {
		return nil
	}
	
	logrus.Infof("Adding peer to network id=%s address=%v", peerID, peerAddresses)
	// add peer to network
	nd.peerAddresses[peerID] = peerAddresses
	for _, addr := range peerAddresses {
		nd.store.Put(getPeerKey(peerID), addr, true)
	}

	// if err := nd.net.PutPeer(peer); err != nil {
	// 	logrus.WithError(err).Warnf("Could not add peer to network")
	// 	return err
	// }
	logrus.Infof("PUT PEER id=%s addrs=%v", peerID, peerAddresses)
	return nil
}

func (nd *DHT) storePeer(peer net.Peer, persistent bool) error {
	for _, addr := range peer.Addresses {
		logrus.WithField("k", getPeerKey(peer.ID)).WithField("v", addr).Infof("Adding peer addresses to kv")
		if err := nd.store.Put(getPeerKey(peer.ID), addr, persistent); err != nil {
			logrus.WithError(err).WithField("peerID", peer.ID).Warnf("storePeer could not put peer")
		}
	}
	return nil
}

// func (nd *DHT) GetLocalPeer() net.Peer {
// 	return nd.localPeer
// }

func (nd *DHT) GetLocalPairs() (map[string][]*Pair, error) {
	return nd.store.GetAll()
}
