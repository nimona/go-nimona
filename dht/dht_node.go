package dht

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	uuid "github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	messagebus "github.com/nimona/go-nimona-messagebus"
	net "github.com/nimona/go-nimona-net"
)

const (
	protocolID       = "dht-kad/v1"
	numPeersNear int = 3
)

type searchEntry struct {
	id              string
	nonce           string
	closestPeer     net.Peer
	shortlistPeers  []net.Peer
	responseChannel chan net.Peer
}

type DHTNode struct {
	// bps are the Bootstrap Peers
	// lp is the local Peer info
	lpeer net.Peer
	// rt is the routing table used
	rt         RoutingTable
	net        net.Network // currently not used
	messageBus messagebus.MessageBus
	// lc stores the nonces and the response channels
	searchStore map[string]*searchEntry
	mt          sync.RWMutex
}

func NewDHTNode(bps []net.Peer, localPeer net.Peer, rt RoutingTable, nnet net.Network) (*DHTNode, error) {
	// create messagebud
	mb, err := messagebus.New(protocolID, nnet, localPeer)
	if err != nil {
		return nil, err
	}

	// create dht node
	dhtNode := &DHTNode{
		lpeer:       localPeer,
		rt:          rt,
		net:         nnet,
		messageBus:  mb,
		searchStore: make(map[string]*searchEntry),
	}

	if err := mb.HandleMessage(dhtNode.messageHandler); err != nil {
		return nil, err
	}

	// add bootstrap nodes
	for _, peer := range bps {
		if err := dhtNode.putPeer(peer); err != nil {
			log.WithField("error", err).Error("Failed to add peer to routing table")
		}
		ctx := context.Background()
		go dhtNode.Find(ctx, peer.ID)
	}

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
		// log.WithField("Type", "FIND_NODE").Info(msg.Payload.Creator)
		go nd.findHandler(dhtMsg)
	default:
		log.Info("Call type not implemented")
	}
	return nil
}

func (nd *DHTNode) Ping(context.Context, net.Peer) (net.Peer, error) {
	return net.Peer{}, nil
}

// TODO: Switch to return channel
func (nd *DHTNode) Find(ctx context.Context, id string) (net.Peer, error) {
	// Search local Routing Table for node
	peer, err := nd.rt.Get(id)
	log.Info("Searching for peer with id: ", id)
	// If node is not found locally send a message to nodes
	if err == ErrPeerNotFound {
		// Create the nonce id
		nc, err := uuid.NewUUID()
		if err != nil {
			log.WithError(err).Error("Failed to generate uuid")
			return net.Peer{}, err
		}

		msg := &Message{
			OriginPeer:  nd.lpeer,
			Nonce:       nc.String(),
			QueryPeerID: id,
		}

		// Check peers in local store for distance
		// send message to the X closest peers
		lookupPeers, err := nd.findPeersNear(id, numPeersNear)
		if err != nil {
			log.WithError(err).Error("Failed find peers near")
		}

		nonce := nc.String()
		responseChannel := make(chan net.Peer)
		se := &searchEntry{
			id:              id,
			nonce:           nonce,
			responseChannel: responseChannel,
			closestPeer:     lookupPeers[0], // TODO Might not exist
		}

		for _, p := range lookupPeers {
			err := nd.sendMsgPeer(MESSAGE_TYPE_FIND_NODE, msg, p.ID)
			if err != nil {
				log.WithError(err).WithField(
					"peer",
					p.ID,
				).Error("Failed to send message to peer")
			}
		}
		nd.mt.Lock()
		nd.searchStore[nc.String()] = se
		nd.mt.Unlock()

		select {
		case rp := <-responseChannel:
			if err = nd.putPeer(rp); err != nil {
				log.Error("Failed to update result peer")
			}
			return rp, nil

		case <-time.After(time.Second * 3):
			return net.Peer{}, ErrPeerNotFound

		case <-ctx.Done():
			return net.Peer{}, ErrPeerNotFound // TODO Better error (context deadline exceeded)
		}

	}
	if err != nil {
		log.WithError(err).Error("Failed to find peer")
		return net.Peer{}, err
	}
	return peer, nil
}

func (nd *DHTNode) putPeer(peer net.Peer) error {
	log.Infof("Adding peer to network peer=%s", peer.ID)
	if err := nd.net.PutPeer(peer); err != nil {
		log.WithError(err).Warnf("Could not add peer to network")
	}
	if err := nd.rt.Save(peer); err != nil {
		return err
	}
	return nil
}

func (nd *DHTNode) sendMsgPeer(msgType string, msg *Message, peerID string) error {
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
	fmt.Println(">", msg.OriginPeer.ID, nd.lpeer.ID)
	// Check if local peer is the origin peer in the message
	if msg.OriginPeer.ID == nd.lpeer.ID {
		nd.mt.Lock()
		defer nd.mt.Unlock()
		if searchEntry, ok := nd.searchStore[msg.Nonce]; ok {
			// Add peers to local routing table
			for _, p := range msg.Peers {
				nd.putPeer(p) // TODO Handle error
			}

			// Check if the requested peer is in the results
			for _, p := range msg.Peers {
				nd.putPeer(p) // TODO Validate peer, handle error
				fmt.Println("LFP", msg.QueryPeerID, p.ID)
				if msg.QueryPeerID == p.ID {
					fmt.Println("FOUND IT", p.ID)
					searchEntry.responseChannel <- p
					// Delete response channel entry
					delete(nd.searchStore, msg.Nonce)
					return
				}
			}

			// Send the request to returned closest peers
			for _, p := range msg.Peers {
				fmt.Println("ADDING", p.ID)
				nd.putPeer(p) // TODO Validate peer, handle error
				err := nd.sendMsgPeer(MESSAGE_TYPE_FIND_NODE, msg, p.ID)
				if err != nil {
					if err != nil {
						log.WithError(err).WithField(
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
		// log.Error("Failed to find node")
	}

	peers, err = nd.findPeersNear(msg.QueryPeerID, numPeersNear)
	if err != nil {
		log.WithField("Msg", msg).Error("Failed to find nodes near")
	}
	if peer.ID != "" && peer.Addresses[0] != "" {
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

// Xor gets to byte arrays and returns and array of integers with the xor
// for between the two equivalent bytes
func xor(a, b []byte) []int {
	compA := []byte{}
	compB := []byte{}
	res := []int{}

	lenA := len(a)
	lenB := len(b)

	// Make both byte arrays have the same size
	if lenA > lenB {
		compA = a
		compB = make([]byte, lenA)
		// Need to leave leftmost bytes empty in order compare
		// the equivalent bytes
		copy(compB[lenA-lenB:], b)
	} else {
		compB = b
		compA = make([]byte, lenB)
		copy(compA[lenB-lenA:], a)
	}

	for i := range compA {
		res = append(res, int(compA[i]^compB[i]))
	}

	return res
}

// distEntry is used to hold the distance between nodes
type distEntry struct {
	id   string
	dist []int
}

// lessIntArr compares two int array return true if a less than b
func lessIntArr(a, b []int) bool {
	for i := range a {
		if a[i] > b[i] {
			return false
		}
		if a[i] < b[i] {
			return true
		}
	}

	return true
}

// findPeersNear accepts an ID and n and finds the n closest nodes to this id
// in the routing table
func (nd *DHTNode) findPeersNear(id string, n int) ([]net.Peer, error) {
	peers := []net.Peer{}

	ids, err := nd.rt.GetPeerIDs()
	if err != nil {
		log.WithError(err).Error("Failed to get peer ids from the routing table")
		return peers, err
	}

	// slice to hold the distances
	dists := []distEntry{}
	for _, pid := range ids {
		entry := distEntry{
			id:   pid,
			dist: xor([]byte(id), []byte(pid)),
		}
		dists = append(dists, entry)
	}
	// Sort the distances
	sort.Slice(dists, func(i, j int) bool {
		return lessIntArr(dists[i].dist, dists[j].dist)
	})

	if n > len(dists) {
		n = len(dists)
	}
	// Append n the first n number of peers from the ids
	for _, de := range dists[:n] {
		p, err := nd.rt.Get(de.id)
		if err != nil {
			log.WithError(err).WithField("ID", de.id).Error("Peer not found")
		}
		peers = append(peers, p)
	}
	return peers, nil
}
