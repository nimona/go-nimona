package dht

import (
	"github.com/mitchellh/mapstructure"
	"nimona.io/go/peers"
	"nimona.io/go/primitives"
)

type PeerInfoResponse struct {
	RequestID    string                `json:"requestID,omitempty" mapstructure:"requestID,omitempty"`
	PeerInfo     *peers.PeerInfo       `json:"peerInfo,omitempty" mapstructure:"peerInfo,omitempty"`
	ClosestPeers []*peers.PeerInfo     `json:"closestPeers,omitempty" mapstructure:"closestPeers,omitempty"`
	Signature    *primitives.Signature `json:"-"`
}

func (r *PeerInfoResponse) Block() *primitives.Block {
	closestPeers := []*primitives.Block{}
	for _, cp := range r.ClosestPeers {
		closestPeers = append(closestPeers, cp.Block())
	}
	return &primitives.Block{
		Type: "nimona.io/dht.peer-info.response",
		Payload: map[string]interface{}{
			"requestID":    r.RequestID,
			"peerInfo":     r.PeerInfo.Block(),
			"closestPeers": closestPeers,
		},
		Signature: r.Signature,
	}
}

func (r *PeerInfoResponse) FromBlock(block *primitives.Block) {
	t := &struct {
		RequestID    string              `mapstructure:"requestID,omitempty"`
		PeerInfo     *primitives.Block   `mapstructure:"peerInfo,omitempty"`
		ClosestPeers []*primitives.Block `mapstructure:"closestPeers,omitempty"`
	}{}

	mapstructure.Decode(block.Payload, t)

	if t.PeerInfo != nil {
		r.PeerInfo = &peers.PeerInfo{}
		r.PeerInfo.FromBlock(t.PeerInfo)
	}

	if len(t.ClosestPeers) > 0 {
		r.ClosestPeers = []*peers.PeerInfo{}
		for _, pb := range t.ClosestPeers {
			pi := &peers.PeerInfo{}
			pi.FromBlock(pb)
			r.ClosestPeers = append(r.ClosestPeers, pi)
		}
	}

	r.RequestID = t.RequestID
	r.Signature = block.Signature
}
