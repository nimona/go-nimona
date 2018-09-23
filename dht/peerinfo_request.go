package dht

import (
	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
)

func init() {
	blocks.RegisterContentType(&PeerInfoRequest{})
}

// PeerInfoRequest payload
type PeerInfoRequest struct {
	RequestID string            `json:"requestID,omitempty"`
	PeerID    string            `json:"peerID"`
	Signature *crypto.Signature `json:"-"`
}

func (p *PeerInfoRequest) GetType() string {
	return "dht.peerinfo.request"
}

func (p *PeerInfoRequest) GetSignature() *crypto.Signature {
	return p.Signature
}

func (p *PeerInfoRequest) SetSignature(s *crypto.Signature) {
	p.Signature = s
}

func (p *PeerInfoRequest) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (p *PeerInfoRequest) SetAnnotations(a map[string]interface{}) {
	// no annotations
}
