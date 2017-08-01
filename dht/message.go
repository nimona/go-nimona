package dht

import (
	net "github.com/nimona/go-nimona-net"
)

const (
	MESSAGE_TYPE_PING       string = "PING"
	MESSAGE_TYPE_STORE             = "STORE"
	MESSAGE_TYPE_FIND_NODE         = "FIND_NODE"
	MESSAGE_TYPE_FIND_VALUE        = "FIND_VALUE"
)

type Message struct {
	Nonce       string     `json:"n"`
	OriginPeer  net.Peer   `json:"op"`
	QueryPeerID string     `json:"qp"`
	Peers       []net.Peer `json:"rp"`
}
