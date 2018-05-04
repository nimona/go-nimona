package dht

import (
	"github.com/nimona/go-nimona/mesh"
)

// Message types
const (
	PayloadTypePing string = "ping"
	PayloadTypePong        = "pong"

	PayloadTypeGetPeerInfo = "get-peer-info"
	PayloadTypePutPeerInfo = "put-peer-info"

	PayloadTypeGetProviders = "get-providers"
	PayloadTypePutProviders = "put-providers"

	PayloadTypeGetValue = "get-value"
	PayloadTypePutValue = "put-value"
)

type messageSenderPeerInfo struct {
	SenderPeerInfo mesh.PeerInfo `json:"sender_peer_info"`
}

type messagePing struct {
	SenderPeerInfo mesh.PeerInfo `json:"sender_peer_info"`
	RequestID      string        `json:"request_id,omitempty"`
	PeerID         string        `json:"peer_id"`
}

type messagePong struct {
	SenderPeerInfo mesh.PeerInfo `json:"sender_peer_info"`
	RequestID      string        `json:"request_id,omitempty"`
	PeerID         string        `json:"peer_id"`
}

type messageGetPeerInfo struct {
	SenderPeerInfo mesh.PeerInfo `json:"sender_peer_info"`
	RequestID      string        `json:"request_id,omitempty"`
	PeerID         string        `json:"peer_id"`
}

type messagePutPeerInfo struct {
	SenderPeerInfo mesh.PeerInfo `json:"sender_peer_info"`
	RequestID      string        `json:"request_id,omitempty"`
	PeerID         string        `json:"peer_id"`
	PeerInfo       mesh.PeerInfo `json:"peer_info"`
	ClosestPeers   []string      `json:"closest_peers"`
}

type messageGetProviders struct {
	SenderPeerInfo mesh.PeerInfo `json:"sender_peer_info"`
	RequestID      string        `json:"request_id,omitempty"`
	Key            string        `json:"key"`
}

type messagePutProviders struct {
	SenderPeerInfo mesh.PeerInfo `json:"sender_peer_info"`
	RequestID      string        `json:"request_id,omitempty"`
	Key            string        `json:"key"`
	PeerIDs        []string      `json:"peer_ids"`
	ClosestPeers   []string      `json:"closest_peers"`
}

type messageGetValue struct {
	SenderPeerInfo mesh.PeerInfo `json:"sender_peer_info"`
	RequestID      string        `json:"request_id,omitempty"`
	Key            string        `json:"key"`
}

type messagePutValue struct {
	SenderPeerInfo mesh.PeerInfo `json:"sender_peer_info"`
	RequestID      string        `json:"request_id,omitempty"`
	Key            string        `json:"key"`
	Value          string        `json:"value"`
	ClosestPeers   []string      `json:"closest_peers"`
}
