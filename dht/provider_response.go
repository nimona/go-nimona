package dht

import (
	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
	"nimona.io/go/peers"
)

func init() {
	blocks.RegisterContentType(&ProviderResponse{})
}

type ProviderResponse struct {
	RequestID    string            `json:"requestID,omitempty"`
	Providers    []*Provider       `json:"providers,omitempty"`
	ClosestPeers []*peers.PeerInfo `json:"closestPeers,omitempty"`
	Signature    *crypto.Signature `json:"-"`
}

func (p *ProviderResponse) GetType() string {
	return "dht.provider.response"
}

func (p *ProviderResponse) GetSignature() *crypto.Signature {
	return p.Signature
}

func (p *ProviderResponse) SetSignature(s *crypto.Signature) {
	p.Signature = s
}

func (p *ProviderResponse) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (p *ProviderResponse) SetAnnotations(a map[string]interface{}) {
	// no annotations
}
