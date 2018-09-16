package dht

import (
	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
	"nimona.io/go/peers"
)

func init() {
	blocks.RegisterContentType(&PeerInfoResponse{})
}

type PeerInfoResponse struct {
	RequestID    string            `json:"requestID,omitempty"`
	PeerInfo     *peers.PeerInfo   `json:"peerInfo,omitempty"`
	ClosestPeers []*peers.PeerInfo `json:"closestPeers,omitempty"`
	Signature    *crypto.Signature `json:"-"`
}

func (p *PeerInfoResponse) GetType() string {
	return "dht.peerinfo.response"
}

func (p *PeerInfoResponse) GetSignature() *crypto.Signature {
	return p.Signature
}

func (p *PeerInfoResponse) SetSignature(s *crypto.Signature) {
	p.Signature = s
}

func (p *PeerInfoResponse) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (p *PeerInfoResponse) SetAnnotations(a map[string]interface{}) {
	// no annotations
}
