package net

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
	RegisterContentType(PeerInfoContentType, PeerInfoPayload{})
}

// PeerInfo holds the information exchange needs to connect to a remote peer
type PeerInfo struct {
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`

	Block *Block `json:"block"`
}
