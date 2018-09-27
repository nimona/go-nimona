package peers

import (
	"github.com/mitchellh/mapstructure"
	"nimona.io/go/primitives"
)

// PeerInfo holds the information exchange needs to connect to a remote peer
type PeerInfo struct {
	Addresses []string `mapstructure:"addresses"`
	Signature *primitives.Signature
}

func (pi *PeerInfo) Block() *primitives.Block {
	return &primitives.Block{
		Type: "nimona.io/peer.info",
		Payload: map[string]interface{}{
			"addresses": pi.Addresses,
		},
		Signature: pi.Signature,
	}
}

func (pi *PeerInfo) FromBlock(block *primitives.Block) {
	if err := mapstructure.Decode(block.Payload, pi); err != nil {
		panic(err)
	}
	pi.Signature = block.Signature
}

func (pi *PeerInfo) Thumbprint() string {
	return pi.Signature.Key.Thumbprint()
}
