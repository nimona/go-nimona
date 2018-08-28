package peers

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/nimona/go-nimona/blocks"
)

// PrivatePeerInfo is a PeerInfo with an additional PrivateKey
type PrivatePeerInfo struct {
	Key       *blocks.Key       `nimona:"-" json:"key"`
	Addresses []string          `nimona:"addresses" json:"-"`
	Signature *blocks.Signature `nimona:",signature" json:"-"`
}

func (pi *PrivatePeerInfo) Thumbprint() string {
	return pi.GetPeerInfo().Thumbprint()
}

func (pi *PrivatePeerInfo) GetPeerInfo() *PeerInfo {
	ppi := &PeerInfo{
		Addresses: pi.Addresses,
		signWith:  pi.Key,
	}
	// HACK to add signature, we should only be signing this when it changes
	b, _ := blocks.Marshal(ppi, blocks.SignWith(pi.Key))
	uppi, _ := blocks.Unmarshal(b)
	fmt.Println(">>>>>>>>>>>>>>", uppi.(*PeerInfo).Signature.Alg)
	return uppi.(*PeerInfo)
}

// GetPrivateKey returns the private key
func (pi *PrivatePeerInfo) GetPrivateKey() *blocks.Key {
	return pi.Key
}

// func (pi *PrivatePeerInfo) MarshalBlock() ([]byte, error) {
// 	return blocks.Marshal(pi, blocks.SignWith(pi.Key))
// }

// func (pi *PrivatePeerInfo) UnmarshalBlock(bytes []byte) error {
// 	return blocks.UnmarshalInto(bytes, pi)
// }

// GetPublicKey returns the public key
func (pi *PrivatePeerInfo) GetPublicKey() *blocks.Key {
	pk := pi.Key.Materialize().(*ecdsa.PrivateKey).Public().(*ecdsa.PublicKey)
	bpk, err := blocks.NewKey(pk)
	if err != nil {
		panic(err)
	}
	return bpk
}
