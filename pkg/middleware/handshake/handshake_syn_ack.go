package handshake

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/peer"
)

//go:generate $GOBIN/objectify -schema /handshake.syn-ack -type SynAck -in handshake_syn_ack.go -out handshake_syn_ack_generated.go

// HandshakeSynAck is the response in the second leg of our net handshake
type SynAck struct {
	Nonce     string            `json:"nonce:s"`
	Peer      *peer.Peer        `json:"peer:o,omitempty"`
	Signature *crypto.Signature `json:"@signature:o"`
}
