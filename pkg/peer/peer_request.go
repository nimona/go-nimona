package peer

import (
	"nimona.io/pkg/crypto"
)

//go:generate $GOBIN/objectify -schema nimona.io/discovery/peer.request -type PeerRequest -in peer_request.go -out peer_request_generated.go

// PeerRequest is a request for a peer info
type PeerRequest struct {
	Keys         []crypto.Fingerprint `json:"keys:ao"`
	ContentTypes []string             `json:"contentTypes:as"`
}
