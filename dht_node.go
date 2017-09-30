package dht

import (
	"context"
	"encoding/json"
	"sync"
	"time"

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
	localPeer    net.Peer
	routingtable *RoutingTable
	net          net.Network
	messageBus   messagebus.MessageBus
	queries      map[string]*query
	mt           sync.RWMutex
}

func NewDHTNode(bps []net.Peer, localPeer net.Peer, routingtable *RoutingTable, nnet net.Network) (*DHTNode, error) {
	// create messagebud
	mb, err := messagebus.New(protocolID, nnet)
	if err != nil {
		return nil, err
	}

	// Create DHT node
	nd := &DHTNode{
		localPeer:    localPeer,
		routingtable: routingtable,
		net:          nnet,
		messageBus:   mb,
		queries:      map[string]*query{},
	}

	// Register message bus, message handler
	if err := mb.HandleMessage(nd.messageHandler); err != nil {
		return nil, err
	}

	// Add bootstrap nodes
	for _, peer := range bps {
		if err := nd.putPeer(peer); err != nil {
			logrus.WithField("error", err).Error("Failed to add peer to routing table")
		}
	}

	// start refresh worker
	ticker := time.NewTicker(15 * time.Second)
	quit := make(chan struct{})
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

	return nd, nil
}

func (nd *DHTNode) refresh() {
	logrus.Infof("Refreshing")
	peers, err := nd.routingtable.FindPeersNear(nd.GetLocalPeer().ID, 25)
	if err != nil {
		logrus.WithError(err).Warnf("refresh could not get peers")
		return
	}
	ctx := context.Background()
	for _, peer := range peers {
		nd.findNode(ctx, peer.ID)
	}
}

func (nd *DHTNode) messageHandler(hash []byte, msg messagebus.Message) error {
	switch msg.Payload.Type {
	case MESSAGE_TYPE_FIND_NODE_REQ:
		dhtMsg := &findNodeRequest{}
		if err := json.Unmarshal(msg.Payload.Data, dhtMsg); err != nil {
			return err
		}
		nd.findNodeRequestHandler(dhtMsg)
	case MESSAGE_TYPE_FIND_NODE_RESP:
		dhtMsg := &findNodeResponse{}
		if err := json.Unmarshal(msg.Payload.Data, dhtMsg); err != nil {
			return err
		}
		nd.findNodeResponseHandler(dhtMsg)
	default:
		logrus.Info("Call type not implemented")
	}
	return nil
}

func (nd *DHTNode) Ping(context.Context, net.Peer) (net.Peer, error) {
	return net.Peer{}, nil
}

// TODO: Switch to return channel
func (nd *DHTNode) Find(ctx context.Context, id string) (net.Peer, error) {
	logrus.Info("Searching for peer with id: ", id)

	// make query
	q, _ := nd.findNode(ctx, id)

	// wait until something happens
	peer, ok := <-q.results

	// return peer not found
	if !ok {
		return peer, ErrPeerNotFound
	}

	return peer, nil
}

func (nd *DHTNode) findNode(ctx context.Context, id string) (*query, error) {
	// create query
	// TODO query needs the context
	query := &query{
		id:                uuid.New().String(),
		dht:               nd,
		peerID:            id,
		contactedPeers:    []string{},
		results:           make(chan net.Peer),
		incomingResponses: make(chan findNodeResponse, 10),
	}

	// and store it
	nd.mt.Lock()
	nd.queries[query.id] = query
	nd.mt.Unlock()

	// run query
	query.Run(ctx)

	return query, nil
}

func (nd *DHTNode) putPeer(peer net.Peer) error {
	if len(peer.Addresses) == 0 {
		return nil
	}

	logrus.Infof("Adding peer to network peer=%v", peer)

	// update peer table
	if err := nd.routingtable.Put(peer); err != nil {
		return err
	}

	// add peer to network
	if err := nd.net.PutPeer(peer); err != nil {
		logrus.WithError(err).Warnf("Could not add peer to network")
		return err
	}

	return nil
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

func (nd *DHTNode) findNodeRequestHandler(req *findNodeRequest) {
	// origin peer is asking for a peer

	// add origin peer to our table
	nd.putPeer(req.OriginPeer)

	// find nearest peers
	peers, err := nd.routingtable.FindPeersNear(req.QueryPeerID, numPeersNear)
	if err != nil {
		logrus.WithField("req", req).Error("Failed to find nodes near")
	}
	// give up if there are no peers
	if len(peers) == 0 {
		return
	}

	// create a response with the peers we found
	res := &findNodeResponse{
		QueryID:     req.QueryID,
		OriginPeer:  req.OriginPeer,
		QueryPeerID: req.QueryPeerID,
		Peers:       peers,
	}

	// send response
	nd.sendMsgPeer(MESSAGE_TYPE_FIND_NODE_RESP, res, req.OriginPeer.ID)
}

func (nd *DHTNode) findNodeResponseHandler(res *findNodeResponse) {
	// A peer we asked is informing us of a peer
	logrus.WithField("queryID", res.QueryID).Infof("Got response")

	// check if this still a valid query
	q, ok := nd.queries[res.QueryID]
	if !ok {
		logrus.WithField("queryID", res.QueryID).Infof("Query no longer exists")
		return
	}

	// send response to query
	q.incomingResponses <- *res
}

func (nd *DHTNode) GetLocalPeer() net.Peer {
	return nd.localPeer
}
