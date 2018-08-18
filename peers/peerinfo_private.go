package peers

import (
	"github.com/nimona/go-nimona/blocks"
	"github.com/nimona/go-nimona/keys"
)

// PrivatePeerInfo is a PeerInfo with an additional PrivateKey
type PrivatePeerInfo struct {
	ID         string   `json:"id"`
	PrivateKey string   `json:"private_key"`
	Addresses  []string `json:"-"`
}

// Block returns a signed Block
func (pi *PrivatePeerInfo) Block() *blocks.Block {
	// TODO content type
	block := blocks.NewEphemeralBlock(PeerInfoType, PeerInfoPayload{
		Addresses: pi.Addresses,
	})
	block.Metadata.Ephemeral = true
	block.Metadata.Signer = pi.ID
	return block
}

// GetPrivateKey returns the private key
func (pi *PrivatePeerInfo) GetPrivateKey() keys.Key {
	sk, err := keys.KeyFromEncodedBlock(pi.PrivateKey)
	if err != nil {
		panic(err)
	}

	return sk
}

// GetPublicKey returns the public key
func (pi *PrivatePeerInfo) GetPublicKey() keys.Key {
	pk, err := keys.KeyFromEncodedBlock(pi.ID)
	if err != nil {
		panic(err)
	}

	return pk
}
