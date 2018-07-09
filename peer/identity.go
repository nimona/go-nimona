package peer

type Identity struct {
	ID        string              `json:"id"`
	Version   int                 `json:"version"`
	PublicKey [32]byte            `json:"public_key"`
	Signature []byte              `json:"signature"`
	Peers     *PeerInfoCollection `json:"-"`
}
