package peers

import (
	"crypto/ecdsa"

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
	// TODO(geoah) not sure if this is needed any more, I think it is
	b, _ := blocks.Marshal(ppi, blocks.SignWith(pi.Key))
	uppi, _ := blocks.Unmarshal(b)
	return uppi.(*PeerInfo)
}

// GetPrivateKey returns the private key
func (pi *PrivatePeerInfo) GetPrivateKey() *blocks.Key {
	return pi.Key
}

// GetPublicKey returns the public key
func (pi *PrivatePeerInfo) GetPublicKey() *blocks.Key {
	pk := pi.Key.Materialize().(*ecdsa.PrivateKey).Public().(*ecdsa.PublicKey)
	bpk, err := blocks.NewKey(pk)
	if err != nil {
		panic(err)
	}
	return bpk
}
