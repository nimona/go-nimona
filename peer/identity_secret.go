package peer

type SecretIdentity struct {
	ID         string              `json:"id"`
	Version    int                 `json:"version"`
	PublicKey  [32]byte            `json:"public_key"`
	SigningKey [64]byte            `json:"signing_key"`
	Signature  []byte              `json:"signature"`
	Peers      *PeerInfoCollection `json:"-"`
}
