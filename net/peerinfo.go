package net

const (
	PeerInfoContentType = "peer.info"
)

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

	Message *Message `json:"message"`
}
