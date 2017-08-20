package dht

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	uuid "github.com/google/uuid"

	messagebus "github.com/nimona/go-nimona-messagebus"
	net "github.com/nimona/go-nimona-net"
)

const (
	protocolID       = "dht-kad/v1"
	numPeersNear int = 3
)

type DHTNode struct {
	// bps are the Bootstrap Peers
	// lp is the local Peer info
	lpeer net.Peer
	// rt is the routing table used
	rt         RoutingTable
	net        net.Network // currently not used
	messageBus messagebus.MessageBus
	// lc stores the nonces and the response channels
	queryStore map[string]*query
	mt         sync.RWMutex
}

func NewDHTNode(bps []net.Peer, localPeer net.Peer, rt RoutingTable, nnet net.Network) (*DHTNode, error) {
	// create messagebud
	mb, err := messagebus.New(protocolID, nnet, localPeer)
	if err != nil {
		return nil, err
	}

	// create dht node
	dhtNode := &DHTNode{
		lpeer:      localPeer,
		rt:         rt,
		net:        nnet,
		messageBus: mb,
		queryStore: make(map[string]*query),
	}

	if err := mb.HandleMessage(dhtNode.messageHandler); err != nil {
		return nil, err
	}

	// add bootstrap nodes
	for _, peer := range bps {
		if err := dhtNode.putPeer(peer); err != nil {
			logrus.WithField("error", err).Error("Failed to add peer to routing table")
		}
		// ctx := context.Background()
		// go dhtNode.Find(ctx, peer.ID)
	}

	go func() {
		// TODO Wait for network
		time.Sleep(time.Second * 2)
		ctx, _ := context.WithTimeout(
			context.Background(),
			time.Second*5,
		)
		if _, err := dhtNode.Find(ctx, localPeer.ID); err != nil {
			logrus.WithError(err).Warnf("Could not find peer %s on startup", localPeer.ID)
		}
	}()

	// Refresh all peers in the routing table
	return dhtNode, nil
}

func (nd *DHTNode) messageHandler(hash []byte, msg messagebus.Message) error {
	switch msg.Payload.Type {
	case MESSAGE_TYPE_FIND_NODE:
		dhtMsg := &Message{}
		if err := json.Unmarshal(msg.Payload.Data, dhtMsg); err != nil {
			return err
		}
		nd.putPeer(dhtMsg.OriginPeer)
		// logrus.WithField("Type", "FIND_NODE").Info(msg.Payload.Creator)
		nd.findHandler(dhtMsg)
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
	// Search local Routing Table for node
	logrus.Info("Searching for peer with id: ", id)
	peer, err := nd.rt.Get(id)
	// If node is not found locally send a message to nodes
	if err != nil {
		if err != ErrPeerNotFound {
			logrus.WithError(err).Error("Failed to find peer")
			return net.Peer{}, err
		}
	}

	logrus.Debugf("Peer does not exist locally")
	nc, err := uuid.NewUUID()
	if err != nil {
		logrus.WithError(err).Error("Failed to generate uuid")
		return net.Peer{}, err
	}

	msg := &Message{
		OriginPeer:  nd.lpeer,
		Nonce:       nc.String(),
		QueryPeerID: id,
	}

	logrus.Debugf("Getting nearest peers")
	// Check peers in local store for distance
	// send message to the X closest peers
	lookupPeers, err := nd.rt.FindPeersNear(id, numPeersNear)
	if err != nil {
		logrus.WithError(err).Error("Failed find peers near")
	}
	logrus.WithField("peers", lookupPeers).Debugf("Asking nearest peers")

	for _, p := range lookupPeers {
		logrus.
			WithField("peer", p.ID).
			WithField("addr", p.Addresses).
			Infof("Asking peer for %s", id)
		err := nd.sendMsgPeer(MESSAGE_TYPE_FIND_NODE, msg, p.ID)
		if err != nil {
			logrus.WithError(err).WithField(
				"peer",
				p.ID,
			).Error("Failed to send message to peer")
		}
	}

	// Store the closest peer
	closestPeer := net.Peer{}
	if len(lookupPeers) > 0 {
		closestPeer = lookupPeers[0]
	}

	nonce := nc.String()
	responseChannel := make(chan net.Peer)
	query := &query{
		id:              id,
		nonce:           nonce,
		responseChannel: responseChannel,
		closestPeer:     closestPeer,
	}
	nd.mt.Lock()
	nd.queryStore[nonce] = query // TODO Check if exists
	nd.mt.Unlock()

	select {
	case rp := <-responseChannel:
		logrus.WithField("rp", rp).Debugf("Found peer, returning")
		if err = nd.putPeer(rp); err != nil {
			logrus.Error("Failed to update result peer")
		}
		return rp, nil

	case <-time.After(time.Second * 3):
		logrus.Warnf("Time has passed")
		return net.Peer{}, ErrPeerNotFound

	case <-ctx.Done():
		logrus.Warnf("CTX was done")
		return net.Peer{}, ErrPeerNotFound // TODO Better error (context deadline exceeded)
	}

	return peer, nil
}

func (nd *DHTNode) putPeer(peer net.Peer) error {
	if len(peer.Addresses) == 0 {
		return nil
	}

	logrus.Infof("Adding peer to network peer=%v", peer)

	// update peer table
	if err := nd.rt.Save(peer); err != nil {
		return err
	}

	// add peer to network
	if err := nd.net.PutPeer(peer); err != nil {
		logrus.WithError(err).Warnf("Could not add peer to network")
		return err
	}

	return nil
}

func (nd *DHTNode) sendMsgPeer(msgType string, msg *Message, peerID string) error {
	if peerID == nd.lpeer.ID {
		return nil
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	pl := &messagebus.Payload{
		Creator: nd.lpeer.ID,
		Coded:   "json",
		Type:    msgType,
		Data:    data,
	}

	return nd.messageBus.Send(pl, []string{peerID})
}

func (nd *DHTNode) findHandler(msg *Message) {
	peers := []net.Peer{}
	rPeers := []net.Peer{}
	// Check if local peer is the origin peer in the message
	if msg.OriginPeer.ID == nd.lpeer.ID {
		nd.mt.Lock()
		defer nd.mt.Unlock()

		if query, ok := nd.queryStore[msg.Nonce]; ok {
			// Check if the requested peer is in the results
			for _, p := range msg.Peers {
				nd.putPeer(p) // TODO Validate peer, handle error
				if msg.QueryPeerID == p.ID {
					fmt.Println("FOUND PEER", p.ID)
					query.responseChannel <- p
					// TODO: Return more than one
					delete(nd.queryStore, msg.Nonce)
					return
				}
			}
			// loopClosestPeer := net.Peer{}
			// Send the request to returned closest peers
			for _, p := range msg.Peers {
				fmt.Println("ADDING", p.ID)
				// TODO Validate peer
				if err := nd.putPeer(p); err != nil {
					logrus.WithError(err).WithField(
						"peer", p.ID,
					).Error("Failed to store peer")
				}
				// Compare if the peer is closer that the current closest
				if comparePeers(p.ID, nd.queryStore[msg.Nonce].closestPeer.ID,
					nd.queryStore[msg.Nonce].id,
				) == p.ID {
					// loopClosestPeer = p
				}
				err := nd.sendMsgPeer(MESSAGE_TYPE_FIND_NODE, msg, p.ID)
				if err != nil {
					if err != nil {
						logrus.WithError(err).WithField(
							"peer",
							p.ID,
						).Error("Failed to send message to peer")
					}
				}
			}
			return
		}
	}

	msg.Peers = []net.Peer{}

	peer, err := nd.rt.Get(msg.QueryPeerID)
	if err != nil {
		// TODO Handle errors other than not found
		// logrus.Error("Failed to find node")
	}

	peers, err = nd.rt.FindPeersNear(msg.QueryPeerID, numPeersNear)
	if err != nil {
		logrus.WithField("Msg", msg).Error("Failed to find nodes near")
	}
	if len(peer.Addresses) > 0 && peer.ID != "" {
		rPeers = append(rPeers, peer)
	} else {
		rPeers = peers
	}
	sm := &Message{
		Nonce:       msg.Nonce,
		OriginPeer:  msg.OriginPeer,
		QueryPeerID: msg.QueryPeerID,
		Peers:       rPeers,
	}
	nd.sendMsgPeer(MESSAGE_TYPE_FIND_NODE, sm, msg.OriginPeer.ID)
}
