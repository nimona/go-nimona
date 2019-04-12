package handshake

import (
	"nimona.io/pkg/net/peer"
	"nimona.io/pkg/object"
)

//go:generate go run nimona.io/tools/objectify -schema /handshake.syn -type Syn -in handshake_syn.go -out handshake_syn_generated.go

type Syn struct {
	RawObject *object.Object `json:"@"`
	Nonce     string         `json:"nonce"`
	PeerInfo  *peer.PeerInfo `json:"peerInfo,omitempty"`
}
