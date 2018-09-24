package dht

import (
	"nimona.io/go/primitives"
)

// PeerInfoRequest payload
type PeerInfoRequest struct {
	RequestID string                `json:"requestID,omitempty"`
	PeerID    string                `json:"peerID"`
	Signature *primitives.Signature `json:"-"`
}

func (r *PeerInfoRequest) Block() *primitives.Block {
	return &primitives.Block{
		Type: "nimona.io/dht.peer-info.request",
		Payload: map[string]interface{}{
			"requestID": r.RequestID,
			"peerID":    r.PeerID,
		},
		Signature: r.Signature,
	}
}

func (r *PeerInfoRequest) FromBlock(block *primitives.Block) {
	r.RequestID = block.Payload["requestID"].(string)
	r.PeerID = block.Payload["peerID"].(string)
	r.Signature = block.Signature
}
