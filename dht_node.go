package dht

import (
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

func (nd *DHTNode) Find(id ID) ([]Peer, error) {
	peer, err := nd.rt.Get(id)
	log.Info("Searching for peer with id: ", id)
	if err == ErrPeerNotFound {
		nc, err := uuid.NewUUID()
		if err != nil {
			log.WithField("error", err).Error("Failed to generate uuid")
			return []Peer{}, err
		}

		msg := &Message{
			Type:        FIND_NODE,
			Nonce:       nc.String(),
			OriginPeer:  *nd.lp,
			QueryPeerID: id,
		}
		for _, bootPeer := range nd.bps {
			for _, addr := range bootPeer.Address {
				i, err := nd.nt.SendMessage(*msg, addr)
				if err != nil {
					log.WithField("error", err).Error("Failed to send message")
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
		log.WithField("error", err).Error("Failed to find peer")
		return []Peer{}, err
	}
	return []Peer{peer}, nil
}

func (nd *DHTNode) handleConnection(conn net.Conn) {
	for {
		// TODO: https://golang.org/pkg/bufio/#Reader.ReadLine
		buffer := make([]byte, 1024)
		_, err := conn.Read(buffer)
		if err != nil {
			log.WithField("error", err).Error("Failed to read from comm")
			return
		}

		log.Info("Message received")

		msg := &Message{}
		err = json.Unmarshal(buffer, msg)
		if err != nil {
			log.WithField("error", err).Error("Failed to unmarshall json")
		}

		// Check if originator is localpeer and nonce exists in local memory
		switch msg.Type {
		case PING:
			log.Info(msg.OriginPeer.ID)
		case FIND_NODE:
			log.Info(msg.OriginPeer.ID)
		default:
			log.Info("Call type not implemented")
		}
	}
}
