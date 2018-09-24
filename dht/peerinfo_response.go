package dht

import (
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
	r.RequestID = block.Payload["requestID"].(string)
	r.PeerInfo = &peers.PeerInfo{}
	if peerInfoMap, ok := block.Payload["peerInfo"].(map[string]interface{}); ok {
		peerInfoBlock := primitives.BlockFromMap(peerInfoMap)
		r.PeerInfo.FromBlock(peerInfoBlock)
	}
	if closestPeersMap, ok := block.Payload["closestPeers"].([]map[string]interface{}); ok {
		r.ClosestPeers = []*peers.PeerInfo{}
		for _, closestPeerMap := range closestPeersMap {
			closestPeerBlock := primitives.BlockFromMap(closestPeerMap)
			closestPeer := &peers.PeerInfo{}
			closestPeer.FromBlock(closestPeerBlock)
			r.ClosestPeers = append(r.ClosestPeers, closestPeer)
		}
	}
	r.Signature = block.Signature
}
