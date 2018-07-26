package dht

import (
	"github.com/nimona/go-nimona/net"
)

func init() {
	net.RegisterContentType(PayloadTypePing, BlockPing{})
	net.RegisterContentType(PayloadTypePong, BlockPong{})

	net.RegisterContentType(PayloadTypeGetPeerInfo, BlockGetPeerInfo{})
	net.RegisterContentType(PayloadTypePutPeerInfo, BlockPutPeerInfoFromBlock{})

	net.RegisterContentType(PayloadProviderType, PayloadProvider{})
	net.RegisterContentType(PayloadTypeGetProviders, BlockGetProviders{})
	net.RegisterContentType(PayloadTypePutProviders, BlockPutProviders{})

	net.RegisterContentType(PayloadTypeGetValue, BlockGetValue{})
	net.RegisterContentType(PayloadTypePutValue, BlockPutValue{})
}

// Block types
const (
	PayloadProviderType = "dht.provider"

	PayloadTypePing = "dht.ping"
	PayloadTypePong = "dht.pong"

	PayloadTypeGetPeerInfo = "dht.get-peer-info"
	PayloadTypePutPeerInfo = "dht.put-peer-info"

	PayloadTypeGetProviders = "dht.get-providers"
	PayloadTypePutProviders = "dht.put-providers"

	PayloadTypeGetValue = "dht.get-value"
	PayloadTypePutValue = "dht.put-value"
)

// BlockPing payload
type BlockPing struct {
	// SenderPeerInfo *net.Block `json:"sender_peer_info"`
	RequestID string `json:"request_id,omitempty"`
	PeerID    string `json:"peer_id"`
}

// BlockPong payload
type BlockPong struct {
	// SenderPeerInfo *net.Block `json:"sender_peer_info"`
	RequestID string `json:"request_id,omitempty"`
	PeerID    string `json:"peer_id"`
}

// BlockGetPeerInfo payload
type BlockGetPeerInfo struct {
	// SenderPeerInfo *net.Block `json:"sender_peer_info"`
	RequestID string `json:"request_id,omitempty"`
	PeerID    string `json:"peer_id"`
}

// BlockPutPeerInfoFromBlock payload
type BlockPutPeerInfoFromBlock struct {
	// SenderPeerInfo *net.Block   `json:"sender_peer_info"`
	RequestID    string       `json:"request_id,omitempty"`
	Peer         *net.Block   `json:"peer"`
	ClosestPeers []*net.Block `json:"closest_peers"`
}

// BlockGetProviders payload
type BlockGetProviders struct {
	// SenderPeerInfo *net.Block `json:"sender_peer_info"`
	RequestID string `json:"request_id,omitempty"`
	Key       string `json:"key"`
}

// BlockPutProviders payload
type BlockPutProviders struct {
	// SenderPeerInfo *net.Block   `json:"sender_peer_info"`
	RequestID    string       `json:"request_id,omitempty"`
	Key          string       `json:"key"`
	Providers    []*net.Block `json:"providers"`
	ClosestPeers []*net.Block `json:"closest_peers"`
}

// BlockGetValue payload
type BlockGetValue struct {
	SenderPeerInfo *net.Block `json:"sender_peer_info"`
	RequestID      string     `json:"request_id,omitempty"`
	Key            string     `json:"key"`
}

// BlockPutValue payload
type BlockPutValue struct {
	SenderPeerInfo *net.Block   `json:"sender_peer_info"`
	RequestID      string       `json:"request_id,omitempty"`
	Key            string       `json:"key"`
	Value          string       `json:"value"`
	ClosestPeers   []*net.Block `json:"closest_peers"`
}

type PayloadProvider struct {
	BlockIDs []string `json:"blockIDs"`
}
