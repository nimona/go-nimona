package mesh

import (
	"sync"

	"github.com/keybase/saltpack/basic"
)

type SecretPeerInfo struct {
	sync.RWMutex
	PeerInfo
	SecretKey [32]byte `json:"secret_key"`
}

func (pi *SecretPeerInfo) GetSecretKey() basic.SecretKey {
	return basic.NewSecretKey(&pi.PublicKey, &pi.SecretKey)
}

func (pi *SecretPeerInfo) UpdateAddresses(addresses []string) {
	pi.Lock()
	pi.Addresses = addresses
	pi.Unlock()
}

func (pi *SecretPeerInfo) ToPeerInfo() PeerInfo {
	return PeerInfo{
		ID:        pi.ID,
		Addresses: pi.Addresses,
		PublicKey: pi.PublicKey,
		Signature: pi.Signature,
	}
}

// func Encrypt(raw []byte, peers []string ) (string, error) {
// 	pks:=[]saltpack.BoxPublicKey{}
// 	for _, peer:=range peers {

// 	}
// 	ciphertext, err = saltpack.EncryptArmor62Seal(saltpack.CurrentVersion(), msg, sender, allReceivers, "")

// }
