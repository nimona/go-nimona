package dht

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"net"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

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
	go nt.StartServer(addr, dhtNode.handleConnection)
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
		return <-result, nil
	}
	if err != nil {
		log.WithError(err).Error("Failed to find peer")
		return []Peer{}, err
	}
	return []Peer{peer}, nil
}

func (nd *DHTNode) handleConnection(conn net.Conn) {
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
			go nd.findReceived(msg)
		default:
			log.Info("Call type not implemented")
		}
	}
}

func (nd *DHTNode) findReceived(msg *Message) {
	// Check if local peer is the originator
	// Check

	// If local peer is not the originator
	// find peers with smallest distance in local store and send them back
}

// findPeersNear accepts an ID and n and finds the n closest nodes to this id
// in the routing table
func (nd *DHTNode) findPeersNear(id ID, n int) ([]*Peer, error) {
	peers := make([]*Peer, n)
	ids, err := nd.rt.GetPeerIDs()
	if err != nil {
		log.WithError(err).Error("Failed to get peer ids from the routing table")
		return peers, err
	}

	dists := make(map[ID][]int, len(ids))
	for _, pid := range ids {
		dists[pid] = Xor([]byte(id), []byte(pid))
	}

	//

	return peers, nil
}

// Xor gets to byte arrays and returns and array of integers with the xor
// for between the two equivalent bytes
func Xor(a, b []byte) []int {
	var compA, compB []byte
	var res = []int{}

	lenA := len(a)
	lenB := len(b)

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
