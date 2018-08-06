package dht

import "github.com/nimona/go-nimona/blocks"

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
	RequestID string `json:"request_id,omitempty"`
	PeerID    string `json:"peer_id"`
}

// ProviderRequest payload
type ProviderRequest struct {
	RequestID string `json:"request_id,omitempty"`
	Key       string `json:"key"`
}

type Provider struct {
	BlockIDs []string `json:"blockIDs"`
}
