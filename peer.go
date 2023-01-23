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
		_         string `cborgen:"$type,const=nimona://peer:key"`
		PublicKey PublicKey
	}
	PeerInfo struct {
		_         string     `cborgen:"$type,const=core/node.info"`
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
	return string(DocumentTypePeerKey) + p.PublicKey.String()
}

func ParsePeerKey(pID string) (PeerKey, error) {
	prefix := string(DocumentTypePeerKey)
	if !strings.HasPrefix(pID, prefix) {
		return PeerKey{}, fmt.Errorf("invalid resource id")
	}

	pID = strings.TrimPrefix(pID, prefix)
	key, err := PublicKeyFromBase58(pID)
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
