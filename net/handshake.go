package net

// HandshakeBlock content structure for Handshake content type
type HandshakeBlock struct {
	PeerInfo *Block `json:"peer_info"`
}
