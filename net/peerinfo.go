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

// PeerInfo holds the information messenger needs to connect to a remote peer
type PeerInfo struct {
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`
	// Status    Status   `json:"-"`
	// UpdatedAt       time.Time `json:"updated_at"`
	// LastConnectedAt time.Time `json:"last_connected_at"`

	Envelope *Envelope `json:"envelope"`
}
