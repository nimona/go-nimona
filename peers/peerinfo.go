package peers

import "github.com/nimona/go-nimona/blocks"

const (
	// PeerInfoType is the content type for PeerInfo
	// TODO Needs better name
	PeerInfoType = "peer.info"
)

func init() {
	blocks.RegisterContentType(PeerInfoType, PeerInfo{})
}

// PeerInfo holds the information exchange needs to connect to a remote peer
type PeerInfo struct {
	RequestID string            `nimona:"requestID,header" json:"-"`
	Addresses []string          `nimona:"addresses" json:"addresses"`
	Signature *blocks.Signature `nimona:",signature" json:"signature"`
	signWith  *blocks.Key
}

func (pi *PeerInfo) Thumbprint() string {
	return pi.Signature.Key.Thumbprint()
}

func (pi *PeerInfo) MarshalBlock() (string, error) {
	var bytes []byte
	var err error
	if pi.signWith != nil {
		bytes, err = blocks.Marshal(pi, blocks.SignWith(pi.signWith))
	} else {
		bytes, err = blocks.Marshal(pi)
	}
	if err != nil {
		return "", err
	}

	return blocks.Base58Encode(bytes), nil

}

func (pi *PeerInfo) UnmarshalBlock(b58bytes string) error {
	bytes, err := blocks.Base58Decode(b58bytes)
	if err != nil {
		return err
	}
	return blocks.UnmarshalInto(bytes, pi)
}
