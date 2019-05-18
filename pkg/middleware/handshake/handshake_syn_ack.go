package handshake

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net/peer"
)

//go:generate $GOBIN/objectify -schema /handshake.syn-ack -type SynAck -in handshake_syn_ack.go -out handshake_syn_ack_generated.go

// HandshakeSynAck is the response in the second leg of our net handshake
type SynAck struct {
	Nonce     string            `json:"nonce"`
	PeerInfo  *peer.PeerInfo    `json:"peerInfo,omitempty"`
	Signature *crypto.Signature `json:"@signature"`
}
