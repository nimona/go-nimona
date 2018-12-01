package dht

import (
	"nimona.io/go/crypto"
	"nimona.io/go/encoding"
	"nimona.io/go/peers"
)

//go:generate go run nimona.io/go/cmd/objectify -schema nimona.io/dht/peerinfo.response -type PeerInfoResponse -out peerinfo_response_generated.go

type PeerInfoResponse struct {
	RequestID    string            `json:"requestID,omitempty"`
	PeerInfo     *peers.PeerInfo   `json:"peerInfo,omitempty"`
	ClosestPeers []*peers.PeerInfo `json:"closestPeers,omitempty"`

	RawObject *encoding.Object  `json:"@"`
	Signer    *crypto.Key       `json:"@signer"`
	Authority *crypto.Key       `json:"@authority"`
	Signature *crypto.Signature `json:"@sig:O"`
}
