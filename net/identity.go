package net

type Identity struct {
	ID    string              `json:"id"`
	Peers *PeerInfoCollection `json:"-"`
}
