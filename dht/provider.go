package dht

import (
	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
)

func init() {
	blocks.RegisterContentType(&Provider{})
}

// Provider payload
type Provider struct {
	BlockIDs  []string          `json:"blockIDs"`
	Signature *crypto.Signature `json:"-"`
}

func (p *Provider) GetType() string {
	return "dht.provider"
}

func (p *Provider) GetSignature() *crypto.Signature {
	return p.Signature
}

func (p *Provider) SetSignature(s *crypto.Signature) {
	p.Signature = s
}

func (p *Provider) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (p *Provider) SetAnnotations(a map[string]interface{}) {
	// no annotations
}
