package dht

import (
	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
)

func init() {
	blocks.RegisterContentType(&ProviderRequest{})
}

// ProviderRequest payload
type ProviderRequest struct {
	RequestID string            `json:"requestID,omitempty"`
	Key       string            `json:"key"`
	Signature *crypto.Signature `json:"-"`
}

func (p *ProviderRequest) GetType() string {
	return "dht.provider.request"
}

func (p *ProviderRequest) GetSignature() *crypto.Signature {
	return p.Signature
}

func (p *ProviderRequest) SetSignature(s *crypto.Signature) {
	p.Signature = s
}

func (p *ProviderRequest) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (p *ProviderRequest) SetAnnotations(a map[string]interface{}) {
	// no annotations
}
