package net // import "nimona.io/go/net"

import (
	"github.com/mitchellh/mapstructure"
	"nimona.io/go/peers"
	"nimona.io/go/primitives"
)

type HandshakeSyn struct {
	Nonce     string                `json:"nonce" mapstructure:"nonce"`
	PeerInfo  *peers.PeerInfo       `json:"peerInfo,omitempty" mapstructure:"peerInfo,omitempty"`
	Signature *primitives.Signature `json:"-"`
}

func (r *HandshakeSyn) Block() *primitives.Block {
	return &primitives.Block{
		Type: "nimona.io/handshake.syn",
		Payload: map[string]interface{}{
			"nonce":    r.Nonce,
			"peerInfo": r.PeerInfo,
		},
		Signature: r.Signature,
	}
}

func (r *HandshakeSyn) FromBlock(block *primitives.Block) {
	t := &struct {
		Nonce    string                 `mapstructure:"nonce"`
		PeerInfo map[string]interface{} `mapstructure:"peerInfo,omitempty"`
	}{}

	mapstructure.Decode(block.Payload, t)

	if t.PeerInfo != nil {
		r.PeerInfo = &peers.PeerInfo{}
		r.PeerInfo.FromBlock(primitives.BlockFromMap(t.PeerInfo))
	}

	r.Nonce = t.Nonce
	r.Signature = block.Signature
}
