package net

type PrivateIdentity struct {
	ID         string              `json:"id"`
	PrivateKey string              `json:"private_key"`
	Peers      *PeerInfoCollection `json:"-"`
}
