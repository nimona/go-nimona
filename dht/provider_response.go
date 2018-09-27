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
	closestPeers := []*primitives.Block{}
	for _, cp := range r.ClosestPeers {
		closestPeers = append(closestPeers, cp.Block())
	}
	providers := []*primitives.Block{}
	for _, p := range r.Providers {
		providers = append(providers, p.Block())
	}
	return &primitives.Block{
		Type: "nimona.io/dht.provider.response",
		Payload: map[string]interface{}{
			"requestID":    r.RequestID,
			"providers":    providers,
			"closestPeers": closestPeers,
		},
		Signature: r.Signature,
	}
}

func (r *ProviderResponse) FromBlock(block *primitives.Block) {
	t := &struct {
		RequestID    string              `mapstructure:"requestID,omitempty"`
		PeerInfo     *primitives.Block   `mapstructure:"peerInfo,omitempty"`
		Providers    []*primitives.Block `mapstructure:"providers,omitempty"`
		ClosestPeers []*primitives.Block `mapstructure:"closestPeers,omitempty"`
	}{}

	mapstructure.Decode(block.Payload, t)

	if len(t.Providers) > 0 {
		r.Providers = []*Provider{}
		for _, pb := range t.Providers {
			pi := &Provider{}
			pi.FromBlock(pb)
			r.Providers = append(r.Providers, pi)
		}
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
