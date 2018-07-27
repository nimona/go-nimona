package net

const (
	TypeHandshake = "handshake"
)

func init() {
	RegisterContentType(TypeHandshake, HandshakeBlock{})
}

// HandshakeBlock content structure for Handshake content type
type HandshakeBlock struct {
	PeerInfo *Block `json:"peer_info"`
}
