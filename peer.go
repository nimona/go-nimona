package nimona

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

func NewPeerKey(key PublicKey) PeerKey {
	return PeerKey{
		PublicKey: key,
	}
}

type (
	PeerKey struct {
		_         string `cborgen:"$type,const=core/peer/key"`
		PublicKey PublicKey
	}
	PeerInfo struct {
		_         string     `cborgen:"$type,const=core/peer/info"`
		Metadata  Metadata   `cborgen:"$metadata,omitempty"`
		PublicKey PublicKey  `cborgen:"publicKey"`
		Addresses []PeerAddr `cborgen:"addresses"`
		RawBytes  []byte     `cborgen:"rawbytes"`
	}
	PeerIdentifier struct {
		PeerKey  *PeerKey
		PeerInfo *PeerInfo
	}
)

func (p PeerKey) String() string {
	return string(ShorthandPeerKey) + p.PublicKey.String()
}

func (p PeerKey) IsZero() bool {
	return p.PublicKey.IsZero()
}

func ParsePeerKey(pID string) (PeerKey, error) {
	prefix := string(ShorthandPeerKey)
	if !strings.HasPrefix(pID, prefix) {
		return PeerKey{}, fmt.Errorf("invalid resource id")
	}

	pID = strings.TrimPrefix(pID, prefix)
	key, err := ParsePublicKey(pID)
	if err != nil {
		return PeerKey{}, fmt.Errorf("invalid public key")
	}

	return PeerKey{PublicKey: key}, nil
}

func (p PeerKey) Value() (driver.Value, error) {
	return p.String(), nil
}

func (p *PeerKey) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if netIDString, ok := value.(string); ok {
		netID, err := ParsePeerKey(netIDString)
		if err != nil {
			return fmt.Errorf("unable to scan into DocumentID: %w", err)
		}
		p.PublicKey = netID.PublicKey
		return nil
	}
	return fmt.Errorf("unable to scan %T into DocumentID", value)
}
