package handshake

import (
	"nimona.io/pkg/peer"
	"nimona.io/pkg/object"
)

//go:generate $GOBIN/objectify -schema /handshake.syn -type Syn -in handshake_syn.go -out handshake_syn_generated.go

type Syn struct {
	RawObject *object.Object `json:"@"`
	Nonce     string         `json:"nonce"`
	Peer  *peer.Peer `json:"peer,omitempty"`
}
