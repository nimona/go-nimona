package nimona

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type (
	NetworkAlias struct {
		_        string `cborgen:"$type,const=core/network/alias"`
		Hostname string
	}
	NetworkIdentity struct {
		_                 string     `cborgen:"$type,const=core/network/id"`
		NetworkInfoRootID DocumentID `cborgen:"networkInfoRootID,omitempty"`
	}
	NetworkInfo struct {
		_             string       `cborgen:"$type,const=core/network/info"`
		Metadata      Metadata     `cborgen:"metadata"`
		NetworkAlias  NetworkAlias `cborgen:"networkAlias"`
		PeerAddresses []PeerAddr   `cborgen:"peerAddresses"`
		RawBytes      []byte       `cborgen:"rawbytes"`
	}
	NetworkIdentifier struct {
		NetworkAlias    *NetworkAlias
		NetworkIdentity *NetworkIdentity
		NetworkInfo     *NetworkInfo
	}
)

func (n NetworkAlias) String() string {
	return string(DocumentTypeNetworkAlias) + n.Hostname
}

func ParseNetworkAlias(handle string) (NetworkAlias, error) {
	t := string(DocumentTypeNetworkAlias)
	if !strings.HasPrefix(handle, t) {
		return NetworkAlias{}, fmt.Errorf("invalid resource id")
	}

	handle = strings.TrimPrefix(handle, t)
	return NetworkAlias{Hostname: handle}, nil
}

func (n NetworkAlias) Value() (driver.Value, error) {
	return n.String(), nil
}

func (n *NetworkAlias) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if netIDString, ok := value.(string); ok {
		netID, err := ParseNetworkAlias(netIDString)
		if err != nil {
			return fmt.Errorf("unable to scan into NetworkAlias: %w", err)
		}
		n.Hostname = netID.Hostname
		return nil
	}
	return fmt.Errorf("unable to scan %T into NetworkAlias", value)
}

func (n NetworkAlias) NetworkIdentifier() NetworkIdentifier {
	return NetworkIdentifier{
		NetworkAlias: &n,
	}
}

func (n NetworkIdentity) NetworkIdentifier() NetworkIdentifier {
	return NetworkIdentifier{
		NetworkIdentity: &n,
	}
}

func (n *NetworkInfo) NetworkIdentifier() NetworkIdentifier {
	return NetworkIdentifier{
		NetworkInfo: n,
	}
}

func (n *NetworkInfo) NetworkIdentity() NetworkIdentity {
	cborBytes := n.RawBytes
	if cborBytes == nil {
		cborBytes, _ = n.MarshalCBORBytes()
	}
	return NetworkIdentity{
		NetworkInfoRootID: NewDocumentIDFromCBOR(cborBytes),
	}
}

func (n NetworkIdentifier) MarshalCBORBytes() ([]byte, error) {
	if n.NetworkAlias != nil {
		return n.NetworkAlias.MarshalCBORBytes()
	}
	if n.NetworkIdentity != nil {
		return n.NetworkIdentity.MarshalCBORBytes()
	}
	if n.NetworkInfo != nil {
		return n.NetworkInfo.MarshalCBORBytes()
	}
	return nil, fmt.Errorf("unable to marshal network identifier")
}

func (n *NetworkIdentifier) UnmarshalCBORBytes(b []byte) error {
	t, err := GetDocumentTypeFromCbor(b)
	if err != nil {
		return fmt.Errorf("unable to find type for network identifier: %w", err)
	}
	switch t {
	case "core/network/alias":
		n.NetworkAlias = &NetworkAlias{}
		return n.NetworkAlias.UnmarshalCBORBytes(b)
	case "core/network/id":
		n.NetworkIdentity = &NetworkIdentity{}
		return n.NetworkIdentity.UnmarshalCBORBytes(b)
	case "core/network/info":
		n.NetworkInfo = &NetworkInfo{}
		return n.NetworkInfo.UnmarshalCBORBytes(b)
	default:
		return fmt.Errorf("unknown network identifier type: %s", t)
	}
}
