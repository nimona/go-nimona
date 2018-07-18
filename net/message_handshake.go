package net

type HandshakeEnvelope struct {
	PeerInfo *Envelope `json:"peer_info"`
}
