package net

// SecretPeerInfo is a PeerInfo with an additional PrivateKey
type SecretPeerInfo struct {
	ID         string   `json:"id"`
	PrivateKey string   `json:"private_key"`
	Addresses  []string `json:"addresses"`
}

// Message returns a signed Message
func (pi *SecretPeerInfo) Message() *Message {
	// TODO content type
	message, _ := NewMessage(PeerInfoContentType, nil, &PeerInfoPayload{
		Addresses: pi.Addresses,
	})
	message.Sign(pi)
	return message
}
