package peers

import (
	"github.com/nimona/go-nimona/blocks"
)

// PrivatePeerInfo is a PeerInfo with an additional PrivateKey
type PrivatePeerInfo struct {
	ID         string   `json:"id"`
	PrivateKey string   `json:"private_key"`
	Addresses  []string `json:"addresses"`
}

// Block returns a signed Block
func (pi *PrivatePeerInfo) Block() *blocks.Block {
	// TODO content type
	block := blocks.NewEphemeralBlock(PeerInfoContentType, PeerInfoPayload{
		Addresses: pi.Addresses,
	})
	block.Metadata.Ephemeral = true
	block.Metadata.Signer = pi.ID
	return block
}
