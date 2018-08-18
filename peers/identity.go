package peers

// Identity is a public identity
type Identity struct {
	ID    string              `json:"id"`
	Peers *PeerInfoCollection `json:"peers"`
}
