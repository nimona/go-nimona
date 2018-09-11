package peers

import (
	"nimona.io/go/blocks"
	"nimona.io/go/crypto"
)

func init() {
	blocks.RegisterContentType(&PeerInfo{})
}

// PeerInfo holds the information exchange needs to connect to a remote peer
type PeerInfo struct {
	Addresses []string          `json:"addresses"`
	Signature *crypto.Signature `json:"-"`
}

func (pi *PeerInfo) GetType() string {
	return "peer.info"
}

func (pi *PeerInfo) GetSignature() *crypto.Signature {
	return pi.Signature
}

func (pi *PeerInfo) SetSignature(s *crypto.Signature) {
	pi.Signature = s
}

func (pi *PeerInfo) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (pi *PeerInfo) SetAnnotations(a map[string]interface{}) {
	// no annotations
}

func (pi *PeerInfo) Thumbprint() string {
	return pi.Signature.Key.Thumbprint()
}
