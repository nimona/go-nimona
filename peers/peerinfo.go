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
	Addresses []string          `nimona:"addresses" json:"addresses"`
	Signature *blocks.Signature `nimona:",signature" json:"signature"`
	signWith  *blocks.Key
}

func (pi *PeerInfo) Thumbprint() string {
	return pi.Signature.Key.Thumbprint()
}

func (pi *PeerInfo) MarshalBlock() ([]byte, error) {
	if pi.signWith != nil {
		return blocks.Marshal(pi, blocks.SignWith(pi.signWith))
	}

	return blocks.Marshal(pi)
}

func (pi *PeerInfo) UnmarshalBlock(bytes []byte) error {
	return blocks.UnmarshalInto(bytes, pi)
}
