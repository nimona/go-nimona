package dht

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/google/uuid"
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

func NewDHTNode(bps []*Peer, localPeer *Peer, rt RoutingTable) *DHTNode {
	nt := &UDPNet{}
	dhtNode := &DHTNode{
		bps: bps,
		lp:  localPeer,
		rt:  rt,
		nt:  nt,
	}

	nt.StartServer(dhtNode.handleConnection)
	return dhtNode
}

func (nd *DHTNode) Find(id ID) ([]Peer, error) {
	peer, err := nd.rt.Get(id)

	if err == ErrPeerNotFound {
		nc, err := uuid.NewUUID()
		if err != nil {
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
				nd.nt.SendMessage(*msg, addr)
			}
		}

		nd.lc[nc.String()] = make(chan []Peer)
		//TODO:  Wait for response in the channel

		return []Peer{peer}, nil
	}
	if err != nil {
		return []Peer{}, err
	}

	return []Peer{peer}, nil
}

func (nd *DHTNode) handleConnection(conn net.Conn) {
	for {
		buffer := make([]byte, 1024)
		i, err := conn.Read(buffer)
		if err != nil {
			return
		}

		fmt.Println("i: ", i, "\tbuffer size: ", len(buffer))

		msg := &Message{}
		err = json.Unmarshal(buffer, msg)
		if err != nil {
			fmt.Println(err)
		}

		// Check if originator is localpeer and nonce exists in local memory
		switch msg.Type {
		case PING:
			fmt.Println(msg.OriginPeer.ID)
		case FIND_NODE:
			fmt.Println(msg.OriginPeer.ID)
		default:
			fmt.Println("Call type not implemented")
		}
	}
}
