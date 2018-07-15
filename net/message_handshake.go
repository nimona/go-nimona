package net

type HandshakeMessage struct {
	PeerInfo *Message `json:"peer_info"`
}
