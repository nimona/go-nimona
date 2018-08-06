package peers

// PrivateIdentity holds the information for a private (local) identity
type PrivateIdentity struct {
	ID         string              `json:"id"`
	PrivateKey string              `json:"private_key"`
	Peers      *PeerInfoCollection `json:"-"`
}
