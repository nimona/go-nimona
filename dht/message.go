package dht

import (
	net "github.com/nimona/go-nimona-net"
)

const (
	MESSAGE_TYPE_PING            string = "PING"
	MESSAGE_TYPE_STORE                  = "STORE"
	MESSAGE_TYPE_FIND_NODE_REQ          = "FIND_NODE_REQ"
	MESSAGE_TYPE_FIND_NODE_RESP         = "FIND_NODE_RESP"
	MESSAGE_TYPE_FIND_VALUE_REQ         = "FIND_VALUE_REQ"
	MESSAGE_TYPE_FIND_VALUE_RESP        = "FIND_VALUE_RESP"
)

type findNodeRequest struct {
	QueryID     string   `json:"n"`
	OriginPeer  net.Peer `json:"op"`
	QueryPeerID string   `json:"qp"`
}

type findNodeResponse struct {
	QueryID     string     `json:"n"`
	OriginPeer  net.Peer   `json:"op"`
	QueryPeerID string     `json:"qp"`
	Peers       []net.Peer `json:"rp"`
}
