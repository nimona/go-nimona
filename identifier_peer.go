package nimona

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

func NewPeerID(key PublicKey) PeerID {
	return PeerID{
		PublicKey: key,
	}
}

type PeerID struct {
	_         string `cborgen:"$prefix,const=nimona://peer:key"`
	PublicKey PublicKey
}

func (p PeerID) String() string {
	return string(ResourceTypePeerKey) + p.PublicKey.String()
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

func (p PeerID) Value() (driver.Value, error) {
	return p.String(), nil
}

func (p *PeerID) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if netIDString, ok := value.(string); ok {
		netID, err := ParsePeerID(netIDString)
		if err != nil {
			return fmt.Errorf("unable to scan into DocumentID: %w", err)
		}
		p.PublicKey = netID.PublicKey
		return nil
	}
	return fmt.Errorf("unable to scan %T into DocumentID", value)
}
