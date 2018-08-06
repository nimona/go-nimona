package peers

import "github.com/nimona/go-nimona/blocks"

const (
	// PeerInfoType is the content type for PeerInfoPayload
	// TODO Needs better name
	PeerInfoType = "peer.info"
)

// PeerInfoPayload content structure for PeerInfoType
type PeerInfoPayload struct {
	Addresses []string `json:"addresses"`
}

func init() {
	blocks.RegisterContentType(PeerInfoType, PeerInfoPayload{})
}

// PeerInfo holds the information exchange needs to connect to a remote peer
type PeerInfo struct {
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`

	Block *blocks.Block `json:"block"`
}
