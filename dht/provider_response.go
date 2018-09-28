package dht

import (
	"github.com/mitchellh/mapstructure"
	"nimona.io/go/peers"
	"nimona.io/go/primitives"
)

type ProviderResponse struct {
	RequestID    string                `json:"requestID,omitempty"`
	Providers    []*Provider           `json:"providers,omitempty"`
	ClosestPeers []*peers.PeerInfo     `json:"closestPeers,omitempty"`
	Signature    *primitives.Signature `json:"-"`
}

func (r *ProviderResponse) Block() *primitives.Block {
	return &primitives.Block{
		Type: "nimona.io/dht.provider.response",
		Payload: map[string]interface{}{
			"requestID":    r.RequestID,
			"providers":    r.Providers,
			"closestPeers": r.ClosestPeers,
		},
		Signature: r.Signature,
	}
}

func (r *ProviderResponse) FromBlock(block *primitives.Block) {
	t := &struct {
		RequestID    string                   `mapstructure:"requestID,omitempty"`
		PeerInfo     *primitives.Block        `mapstructure:"peerInfo,omitempty"`
		Providers    []map[string]interface{} `mapstructure:"providers,omitempty"`
		ClosestPeers []map[string]interface{} `mapstructure:"closestPeers,omitempty"`
	}{}

	mapstructure.Decode(block.Payload, t)

	if len(t.Providers) > 0 {
		r.Providers = []*Provider{}
		for _, pb := range t.Providers {
			pi := &Provider{}
			pi.FromBlock(primitives.BlockFromMap(pb))
			r.Providers = append(r.Providers, pi)
		}
	}

	if len(t.ClosestPeers) > 0 {
		r.ClosestPeers = []*peers.PeerInfo{}
		for _, pb := range t.ClosestPeers {
			pi := &peers.PeerInfo{}
			pi.FromBlock(primitives.BlockFromMap(pb))
			r.ClosestPeers = append(r.ClosestPeers, pi)
		}
	}

	r.RequestID = t.RequestID
	r.Signature = block.Signature
}
