package dht

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/net/peer"
)

//go:generate go run nimona.io/tools/objectify -schema nimona.io/dht/peerinfo.response -type PeerInfoResponse -in peerinfo_response.go -out peerinfo_response_generated.go

type PeerInfoResponse struct {
	RequestID    string            `json:"requestID,omitempty"`
	PeerInfo     *peer.PeerInfo   `json:"peerInfo,omitempty"`
	ClosestPeers []*peer.PeerInfo `json:"closestPeers,omitempty"`

	RawObject *object.Object  `json:"@"`
	Signer    *crypto.Key       `json:"@signer"`
	Authority *crypto.Key       `json:"@authority"`
	Signature *crypto.Signature `json:"@signature"`
}
