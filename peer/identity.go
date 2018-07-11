package peer

type Identity struct {
	ID        string              `json:"id"`
	Version   int                 `json:"version"`
	PublicKey []byte              `json:"public_key"`
	Signature []byte              `json:"signature"`
	Peers     *PeerInfoCollection `json:"-"`
}

type LocalIdentity struct {
	ID         string              `json:"id"`
	Version    int                 `json:"version"`
	PrivateKey []byte              `json:"private_key"`
	PublicKey  []byte              `json:"public_key"`
	Signature  []byte              `json:"signature"`
	Peers      *PeerInfoCollection `json:"-"`
}
