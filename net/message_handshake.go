package net

// HandshakeEnvelope content structure for Handshake content type
type HandshakeEnvelope struct {
	PeerInfo *Envelope `json:"peer_info"`
}
