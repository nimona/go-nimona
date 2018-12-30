package dht

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/encoding"
)

//go:generate go run nimona.io/tools/objectify -schema nimona.io/dht/peerinfo.request -type PeerInfoRequest -in peerinfo_request.go -out peerinfo_request_generated.go

// PeerInfoRequest payload
type PeerInfoRequest struct {
	RequestID string `json:"requestID,omitempty"`
	PeerID    string `json:"peerID"`

	RawObject *encoding.Object  `json:"@"`
	Signer    *crypto.Key       `json:"@signer"`
	Authority *crypto.Key       `json:"@authority"`
	Signature *crypto.Signature `json:"@signature"`
}
