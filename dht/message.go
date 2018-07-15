package dht

import (
	"github.com/nimona/go-nimona/net"
)

func init() {
	net.RegisterContentType(PayloadTypePing, MessagePing{})
	net.RegisterContentType(PayloadTypePong, MessagePong{})

	net.RegisterContentType(PayloadTypeGetPeerInfo, MessageGetPeerInfo{})
	net.RegisterContentType(PayloadTypePutPeerInfo, MessagePutPeerInfoFromMessage{})

	net.RegisterContentType(PayloadTypeGetProviders, MessageGetProviders{})
	net.RegisterContentType(PayloadTypePutProviders, MessagePutProviders{})

	net.RegisterContentType(PayloadTypeGetValue, MessageGetValue{})
	net.RegisterContentType(PayloadTypePutValue, MessagePutValue{})
}

// Message types
const (
	PayloadTypePing string = "dht.ping"
	PayloadTypePong        = "dht.pong"

	PayloadTypeGetPeerInfo = "dht.get-peer-info"
	PayloadTypePutPeerInfo = "dht.put-peer-info"

	PayloadTypeGetProviders = "dht.get-providers"
	PayloadTypePutProviders = "dht.put-providers"

	PayloadTypeGetValue = "dht.get-value"
	PayloadTypePutValue = "dht.put-value"
)

type MessagePing struct {
	SenderPeerInfo *net.Message `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	PeerID         string       `json:"peer_id"`
}

type MessagePong struct {
	SenderPeerInfo *net.Message `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	PeerID         string       `json:"peer_id"`
}

type MessageGetPeerInfo struct {
	SenderPeerInfo *net.Message `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	PeerID         string       `json:"peer_id"`
}

type MessagePutPeerInfoFromMessage struct {
	SenderPeerInfo *net.Message   `json:"sender_peer_info"`
	RequestID      string         `json:"request_id,omitempty"`
	PeerInfo       *net.Message   `json:"peer_info"`
	ClosestPeers   []*net.Message `json:"closest_peers"`
}

type MessageGetProviders struct {
	SenderPeerInfo *net.Message `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	Key            string       `json:"key"`
}

type MessagePutProviders struct {
	SenderPeerInfo *net.Message   `json:"sender_peer_info"`
	RequestID      string         `json:"request_id,omitempty"`
	Key            string         `json:"key"`
	PeerIDs        []string       `json:"peer_ids"`
	ClosestPeers   []*net.Message `json:"closest_peers"`
}

type MessageGetValue struct {
	SenderPeerInfo *net.Message `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	Key            string       `json:"key"`
}

type MessagePutValue struct {
	SenderPeerInfo *net.Message   `json:"sender_peer_info"`
	RequestID      string         `json:"request_id,omitempty"`
	Key            string         `json:"key"`
	Value          string         `json:"value"`
	ClosestPeers   []*net.Message `json:"closest_peers"`
}
