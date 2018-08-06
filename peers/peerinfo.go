package peers

import "github.com/nimona/go-nimona/blocks"

const (
	// PeerInfoContentType is the content type for PeerInfoPayload
	// TODO Needs better name
	PeerInfoContentType = "peer.info"
)

// PeerInfoPayload content structure for PeerInfoContentType
type PeerInfoPayload struct {
	Addresses []string `json:"addresses"`
}

func init() {
	blocks.RegisterContentType(PeerInfoContentType, PeerInfoPayload{})
}

// PeerInfo holds the information exchange needs to connect to a remote peer
type PeerInfo struct {
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`

	Block *blocks.Block `json:"block"`
}
