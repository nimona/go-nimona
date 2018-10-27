package dht // import "nimona.io/go/dht"

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
	return &primitives.Block{
		Type: "nimona.io/dht.peer-info.response",
		Payload: map[string]interface{}{
			"requestID":    r.RequestID,
			"peerInfo":     r.PeerInfo,
			"closestPeers": r.ClosestPeers,
		},
		Signature: r.Signature,
	}
}

func (r *PeerInfoResponse) FromBlock(block *primitives.Block) {
	t := &struct {
		RequestID    string                   `mapstructure:"requestID,omitempty"`
		PeerInfo     map[string]interface{}   `mapstructure:"peerInfo,omitempty"`
		ClosestPeers []map[string]interface{} `mapstructure:"closestPeers,omitempty"`
	}{}

	mapstructure.Decode(block.Payload, t)

	if t.PeerInfo != nil {
		r.PeerInfo = &peers.PeerInfo{}
		r.PeerInfo.FromBlock(primitives.BlockFromMap(t.PeerInfo))
	}

	if len(t.ClosestPeers) > 0 {
		r.ClosestPeers = []*peers.PeerInfo{}
		for _, pb := range t.ClosestPeers {
			pi := &peers.PeerInfo{}
			ppb := primitives.BlockFromMap(pb)
			pi.FromBlock(ppb)
			r.ClosestPeers = append(r.ClosestPeers, pi)
		}
	}

	r.RequestID = t.RequestID
	r.Signature = block.Signature
}
