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

// User A on Peer 1 is looking to talk to User b on any peer with protocol X
// GetProviders users/b, which returns peers 20 and 21
// GetPeerInfo 20, GetPeerInfo21

type messagePing struct {
	RequestID string `json:"request_id,omitempty"`
	PeerID    string `json:"peer_id"`
}

type messagePong struct {
	RequestID string `json:"request_id,omitempty"`
	PeerID    string `json:"peer_id"`
}

type messageGetPeerInfo struct {
	RequestID string `json:"request_id,omitempty"`
	PeerID    string `json:"peer_id"`
}

type messagePutPeerInfo struct {
	RequestID    string        `json:"request_id,omitempty"`
	PeerID       string        `json:"peer_id"`
	PeerInfo     mesh.PeerInfo `json:"peer_info"`
	ClosestPeers []string      `json:"closest_peers"`
}

type messageGetProviders struct {
	RequestID string `json:"request_id,omitempty"`
	Key       string `json:"key"`
}

type messagePutProviders struct {
	RequestID    string   `json:"request_id,omitempty"`
	Key          string   `json:"key"`
	PeerIDs      []string `json:"peer_ids"`
	ClosestPeers []string `json:"closest_peers"`
}

type messageGetValue struct {
	RequestID string `json:"request_id,omitempty"`
	Key       string `json:"key"`
}

type messagePutValue struct {
	RequestID    string   `json:"request_id,omitempty"`
	Key          string   `json:"key"`
	Value        string   `json:"value"`
	ClosestPeers []string `json:"closest_peers"`
}
