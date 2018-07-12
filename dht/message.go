package dht

import (
	"github.com/nimona/go-nimona/net"
)

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

type messageSenderPeerInfo struct {
	SenderPeerInfo net.PeerInfo `json:"sender_peer_info"`
}

type messagePing struct {
	SenderPeerInfo net.PeerInfo `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	PeerID         string       `json:"peer_id"`
}

type messagePong struct {
	SenderPeerInfo net.PeerInfo `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	PeerID         string       `json:"peer_id"`
}

type messageGetPeerInfo struct {
	SenderPeerInfo net.PeerInfo `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	PeerID         string       `json:"peer_id"`
}

type messagePutPeerInfo struct {
	SenderPeerInfo net.PeerInfo    `json:"sender_peer_info"`
	RequestID      string          `json:"request_id,omitempty"`
	PeerID         string          `json:"peer_id"`
	PeerInfo       net.PeerInfo    `json:"peer_info"`
	ClosestPeers   []*net.PeerInfo `json:"closest_peers"`
}

type messageGetProviders struct {
	SenderPeerInfo net.PeerInfo `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	Key            string       `json:"key"`
}

type messagePutProviders struct {
	SenderPeerInfo net.PeerInfo    `json:"sender_peer_info"`
	RequestID      string          `json:"request_id,omitempty"`
	Key            string          `json:"key"`
	PeerIDs        []string        `json:"peer_ids"`
	ClosestPeers   []*net.PeerInfo `json:"closest_peers"`
}

type messageGetValue struct {
	SenderPeerInfo net.PeerInfo `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	Key            string       `json:"key"`
}

type messagePutValue struct {
	SenderPeerInfo net.PeerInfo    `json:"sender_peer_info"`
	RequestID      string          `json:"request_id,omitempty"`
	Key            string          `json:"key"`
	Value          string          `json:"value"`
	ClosestPeers   []*net.PeerInfo `json:"closest_peers"`
}
