package nimona

import (
	"fmt"
	"strings"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
)

type PeerID struct {
	PublicKey ed25519.PublicKey
}

func (p PeerID) String() string {
	return string(ResourceTypePeerKey) + PublicKeyToBase58(p.PublicKey)
}

func ParsePeerID(pID string) (PeerID, error) {
	prefix := string(ResourceTypePeerKey)
	if !strings.HasPrefix(pID, prefix) {
		return PeerID{}, fmt.Errorf("invalid resource id")
	}

	pID = strings.TrimPrefix(pID, prefix)
	key, err := PublicKeyFromBase58(pID)
	if err != nil {
		return PeerID{}, fmt.Errorf("invalid public key")
	}

	return PeerID{PublicKey: key}, nil
}
