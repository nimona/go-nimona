package dht

import (
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
	r.RequestID = block.Payload["requestID"].(string)
	closestPeersMap := block.Payload["closestPeers"].([]map[string]interface{})
	r.ClosestPeers = []*peers.PeerInfo{}
	for _, closestPeerMap := range closestPeersMap {
		closestPeerBlock := primitives.BlockFromMap(closestPeerMap)
		closestPeer := &peers.PeerInfo{}
		closestPeer.FromBlock(closestPeerBlock)
		r.ClosestPeers = append(r.ClosestPeers, closestPeer)
	}
	providersMap := block.Payload["providers"].([]map[string]interface{})
	r.Providers = []*Provider{}
	for _, providerMap := range providersMap {
		providerBlock := primitives.BlockFromMap(providerMap)
		provider := &Provider{}
		provider.FromBlock(providerBlock)
		r.Providers = append(r.Providers, provider)
	}
	r.Signature = block.Signature
}
