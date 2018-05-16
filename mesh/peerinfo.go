package mesh

import "errors"

type PeerInfo struct {
	ID        string   `json:"id"`
	Addresses []string `json:"addresses"`
	PublicKey []byte   `json:"public_key"`
}

func (pi *PeerInfo) IsValid() bool {
	pk := DecocdePublicKey(pi.PublicKey)
	return IDFromPublicKey(*pk) == pi.ID
}

func NewPeerInfo(id string, addresses []string, publicKey []byte) (*PeerInfo, error) {
	if id == "" {
		return nil, errors.New("missing id")
	}

	if len(addresses) == 0 {
		return nil, errors.New("missing addresses")
	}

	if len(publicKey) == 0 {
		return nil, errors.New("missing public key")
	}

	pi := &PeerInfo{
		ID:        id,
		Addresses: addresses,
		PublicKey: publicKey,
	}

	if !pi.IsValid() {
		return nil, errors.New("id and pk don't match")
	}

	return pi, nil
}
