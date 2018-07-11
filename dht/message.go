package dht

import (
	"github.com/nimona/go-nimona/peer"
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
	SenderPeerInfo peer.PeerInfo `json:"sender_peer_info"`
}

type messagePing struct {
	SenderPeerInfo peer.PeerInfo `json:"sender_peer_info"`
	RequestID      string        `json:"request_id,omitempty"`
	PeerID         string        `json:"peer_id"`
}

type messagePong struct {
	SenderPeerInfo peer.PeerInfo `json:"sender_peer_info"`
	RequestID      string        `json:"request_id,omitempty"`
	PeerID         string        `json:"peer_id"`
}

type messageGetPeerInfo struct {
	SenderPeerInfo peer.PeerInfo `json:"sender_peer_info"`
	RequestID      string        `json:"request_id,omitempty"`
	PeerID         string        `json:"peer_id"`
}

type messagePutPeerInfo struct {
	SenderPeerInfo peer.PeerInfo    `json:"sender_peer_info"`
	RequestID      string           `json:"request_id,omitempty"`
	PeerID         string           `json:"peer_id"`
	PeerInfo       peer.PeerInfo    `json:"peer_info"`
	ClosestPeers   []*peer.PeerInfo `json:"closest_peers"`
}

type messageGetProviders struct {
	SenderPeerInfo peer.PeerInfo `json:"sender_peer_info"`
	RequestID      string        `json:"request_id,omitempty"`
	Key            string        `json:"key"`
}

type messagePutProviders struct {
	SenderPeerInfo peer.PeerInfo    `json:"sender_peer_info"`
	RequestID      string           `json:"request_id,omitempty"`
	Key            string           `json:"key"`
	PeerIDs        []string         `json:"peer_ids"`
	ClosestPeers   []*peer.PeerInfo `json:"closest_peers"`
}

type messageGetValue struct {
	SenderPeerInfo peer.PeerInfo `json:"sender_peer_info"`
	RequestID      string        `json:"request_id,omitempty"`
	Key            string        `json:"key"`
}

type messagePutValue struct {
	SenderPeerInfo peer.PeerInfo    `json:"sender_peer_info"`
	RequestID      string           `json:"request_id,omitempty"`
	Key            string           `json:"key"`
	Value          string           `json:"value"`
	ClosestPeers   []*peer.PeerInfo `json:"closest_peers"`
}
