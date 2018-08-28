package dht

import (
	"github.com/nimona/go-nimona/blocks"
)

func init() {
	blocks.RegisterContentType(ProviderType, Provider{})
	blocks.RegisterContentType(PeerInfoRequestType, PeerInfoRequest{})
	blocks.RegisterContentType(ProviderRequestType, ProviderRequest{})
}

// Block types
const (
	ProviderType = "dht.provider"

	PeerInfoRequestType = "dht.get-peer-info"
	ProviderRequestType = "dht.get-providers"
)

// PeerInfoRequest payload
type PeerInfoRequest struct {
	RequestID string            `nimona:",header:requestID" json:"requestID,omitempty"`
	PeerID    string            `nimona:"peerID" json:"peerID"`
	Signature *blocks.Signature `nimona:",signature" json:"signature"`
}

// ProviderRequest payload
type ProviderRequest struct {
	RequestID string            `nimona:"requestID" json:"requestID,omitempty"`
	Key       string            `nimona:"key" json:"key"`
	Signature *blocks.Signature `nimona:",signature" json:"signature"`
}

// Provider payload
type Provider struct {
	BlockIDs  []string          `json:"blockIDs"`
	Signature *blocks.Signature `nimona:",signature" json:"signature"`
}
