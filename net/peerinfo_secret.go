package net

// PrivatePeerInfo is a PeerInfo with an additional PrivateKey
type PrivatePeerInfo struct {
	ID         string   `json:"id"`
	PrivateKey string   `json:"private_key"`
	Addresses  []string `json:"addresses"`
}

// Block returns a signed Block
func (pi *PrivatePeerInfo) Block() *Block {
	// TODO content type
	block := NewBlock(PeerInfoContentType, PeerInfoPayload{
		Addresses: pi.Addresses,
	})
	block.Metadata.Ephemeral = true
	SetSigner(block, pi)
	Sign(block, pi)
	SetID(block)
	return block
}
