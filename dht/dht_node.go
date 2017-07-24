package dht

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"net"
	"sort"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

const numPeersNear int = 3

type DHTNode struct {
	// bps are the Bootstrap Peers
	bps []*Peer
	// lp is the local Peer info
	lp *Peer
	// rt is the routing table used
	rt RoutingTable
	// nt is the network interface used for comms
	nt Net
	// lc stores the nonces and the response channels
	lc map[string]chan []Peer
}

func NewDHTNode(bps []*Peer, localPeer *Peer, rt RoutingTable, addr string) *DHTNode {
	nt := &UDPNet{}
	dhtNode := &DHTNode{
		bps: bps,
		lp:  localPeer,
		rt:  rt,
		nt:  nt,
		lc:  make(map[string]chan []Peer),
	}
	log.WithField("address", addr).Info("Server starting...")
	go nt.StartServer(addr, dhtNode.connectionHandler)
	return dhtNode
}

// TODO: Switch to return channel
func (nd *DHTNode) Find(ctx context.Context, id ID) ([]Peer, error) {
	// Search local Routing Table for node
	peer, err := nd.rt.Get(id)
	log.Info("Searching for peer with id: ", id)
	// If node is not found locally send a message to nodes
	if err == ErrPeerNotFound {
		nc, err := uuid.NewUUID()
		if err != nil {
			log.WithError(err).Error("Failed to generate uuid")
			return []Peer{}, err
		}

		msg := &Message{
			Type:        FIND_NODE,
			Nonce:       nc.String(),
			OriginPeer:  *nd.lp,
			QueryPeerID: id,
		}

		// Check peers in local store for distance
		// send message to the X closest peers

		// If no peers found in local store
		// send message to all bootstrap nodes
		for _, bootPeer := range nd.bps {
			for _, addr := range bootPeer.Address {
				i, err := nd.nt.SendMessage(*msg, addr)
				if err != nil {
					log.WithError(err).Error("Failed to send message")
				}
				log.Info("Sent message: ", i)
			}
		}

		result := make(chan []Peer)
		nd.lc[nc.String()] = result
		// timeout to wait for response
		log.Info("Waiting for response")
		return <-result, nil
	}
	if err != nil {
		log.WithError(err).Error("Failed to find peer")
		return []Peer{}, err
	}
	return []Peer{peer}, nil
}

func (nd *DHTNode) connectionHandler(conn net.Conn) {
	for {
		buffer := make([]byte, 1024)
		_, err := conn.Read(buffer)
		if err != nil {
			log.WithError(err).Error("Failed to read from comm")
			return
		}

		log.Info("Message received")

		msg := &Message{}
		buflen, uvlen := binary.Uvarint(buffer)
		err = json.Unmarshal(buffer[uvlen:uvlen+int(buflen)], msg)
		if err != nil {
			log.WithError(err).Error("Failed to unmarshall json")
		}

		// Check if originator is localpeer and nonce exists in local memory
		switch msg.Type {
		case PING:
			log.WithField("Type", "PING").Info(msg.OriginPeer.ID)
		case FIND_NODE:
			log.WithField("Type", "FIND_NODE").Info(msg.OriginPeer.ID)
			go nd.findHandler(msg)
		default:
			log.Info("Call type not implemented")
		}
	}
}

func (nd *DHTNode) findHandler(msg *Message) {
	var peers []Peer
	rPeers := []Peer{}
	if msg.OriginPeer.ID == nd.lp.ID {
		nd.lc[msg.Nonce] <- msg.Peers
		return
	}

	peer, err := nd.rt.Get(msg.QueryPeerID)
	if err != nil {
		log.Error("Failed to find node")
	}

	peers, err = nd.findPeersNear(msg.QueryPeerID, numPeersNear)
	if err != nil {
		log.WithField("Msg", msg).Error("Failed to find nodes near")
	}
	if peer.ID != "" && peer.Address[0] != "" {
		rPeers = append(rPeers, peer)
	} else {
		rPeers = peers
	}
	nd.nt.SendMessage(
		Message{
			Type:        FIND_NODE,
			Nonce:       msg.Nonce,
			OriginPeer:  msg.OriginPeer,
			QueryPeerID: msg.QueryPeerID,
			Peers:       rPeers,
		},
		msg.OriginPeer.Address[0],
	)
}

// Xor gets to byte arrays and returns and array of integers with the xor
// for between the two equivalent bytes
func xor(a, b []byte) []int {
	var compA, compB []byte
	var res = []int{}

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
	id   ID
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
func (nd *DHTNode) findPeersNear(id ID, n int) ([]Peer, error) {
	peers := []Peer{}

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
