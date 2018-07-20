package net

// PrivatePeerInfo is a PeerInfo with an additional PrivateKey
type PrivatePeerInfo struct {
	ID         string   `json:"id"`
	PrivateKey string   `json:"private_key"`
	Addresses  []string `json:"addresses"`
}

// Envelope returns a signed Envelope
func (pi *PrivatePeerInfo) Envelope() *Envelope {
	// TODO content type
	envelope := NewEnvelope(PeerInfoContentType, nil, PeerInfoPayload{
		Addresses: pi.Addresses,
	})
	envelope.Sign(pi)
	return envelope
}
