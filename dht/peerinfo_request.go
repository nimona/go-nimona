package dht

import (
	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
)

//go:generate go run nimona.io/go/cmd/objectify -schema nimona.io/dht/peerinfo.request -type PeerInfoRequest -out peerinfo_request_generated.go

// PeerInfoRequest payload
type PeerInfoRequest struct {
	RequestID string `json:"requestID,omitempty"`
	PeerID    string `json:"peerID"`

	RawObject *encoding.Object  `json:"@"`
	Signer    *crypto.Key       `json:"@signer"`
	Authority *crypto.Key       `json:"@authority"`
	Signature *crypto.Signature `json:"@sig:O"`
}
