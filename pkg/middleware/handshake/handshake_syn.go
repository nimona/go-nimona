package handshake

import (
	"nimona.io/pkg/peer"
)

//go:generate $GOBIN/objectify -schema /handshake.syn -type Syn -in handshake_syn.go -out handshake_syn_generated.go

type Syn struct {
	Nonce string     `json:"nonce:s"`
	Peer  *peer.Peer `json:"peer:o,omitempty"`
}
