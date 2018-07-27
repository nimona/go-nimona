package dht

import (
	"github.com/nimona/go-nimona/net"
)

func init() {
	net.RegisterContentType(PayloadTypeGetPeerInfo, BlockGetPeerInfo{})
	net.RegisterContentType(PayloadTypePutPeerInfo, BlockPutPeerInfoFromBlock{})

	net.RegisterContentType(PayloadProviderType, PayloadProvider{})
	net.RegisterContentType(PayloadTypeGetProviders, BlockGetProviders{})
	net.RegisterContentType(PayloadTypePutProviders, BlockPutProviders{})
}

// Block types
const (
	PayloadProviderType = "dht.provider"

	PayloadTypeGetPeerInfo = "dht.get-peer-info"
	PayloadTypePutPeerInfo = "dht.put-peer-info"

	PayloadTypeGetProviders = "dht.get-providers"
	PayloadTypePutProviders = "dht.put-providers"
)

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

type PayloadProvider struct {
	BlockIDs []string `json:"blockIDs"`
}
