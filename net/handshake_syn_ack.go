package net

import (
	"nimona.io/go/crypto"
	"nimona.io/go/peers"
)

//go:generate go run nimona.io/go/cmd/objectify -schema /handshake.syn-ack -type HandshakeSynAck -out handshake_syn_ack_generated.go

// HandshakeSynAck is the response in the second leg of our net handshake
type HandshakeSynAck struct {
	Nonce     string            `json:"nonce"`
	PeerInfo  *peers.PeerInfo   `json:"peerInfo,omitempty"`
	Signer    *crypto.Key       `json:"@signer"`
	Signature *crypto.Signature `json:"@signature"`
}
