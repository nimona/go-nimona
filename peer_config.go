package nimona

import (
	"sync"
)

type PeerConfig struct {
	mutex      sync.RWMutex
	privateKey PrivateKey
	publicKey  PublicKey
	peerInfo   *PeerInfo
}

func NewPeerConfig(
	privateKey PrivateKey,
	publicKey PublicKey,
	peerInfo *PeerInfo,
) *PeerConfig {
	return &PeerConfig{
		privateKey: privateKey,
		publicKey:  publicKey,
		peerInfo:   peerInfo,
	}
}

func (pc *PeerConfig) GetPrivateKey() PrivateKey {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return pc.privateKey
}

func (pc *PeerConfig) GetPublicKey() PublicKey {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return pc.publicKey
}

func (pc *PeerConfig) GetPeerInfo() *PeerInfo {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return pc.peerInfo
}
