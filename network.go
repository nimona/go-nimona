package nimona

import (
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
)

type (
	NetworkAlias struct {
		_        string `cborgen:"$type,const=core/network/alias"`
		Hostname string `cborgen:"hostname"`
	}
	NetworkIdentity struct {
		_                 string       `cborgen:"$type,const=core/network"`
		NetworkInfoRootID DocumentID   `cborgen:"networkInfoRootID,omitempty"`
		NetworkAlias      NetworkAlias `cborgen:"networkAlias,omitempty"`
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
	return string(ShorthandNetworkAlias) + n.Hostname
}

func ParseNetworkAlias(handle string) (NetworkAlias, error) {
	t := string(ShorthandNetworkAlias)
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
		cborBytes, _ = MarshalCBORBytes(n)
	}
	return NetworkIdentity{
		NetworkInfoRootID: NewDocumentIDFromCBOR(cborBytes),
	}
}

func (n *NetworkInfo) String() string {
	return n.NetworkAlias.String()
}

func (n *NetworkIdentifier) MarshalCBOR(w io.Writer) error {
	if n.NetworkAlias != nil {
		return n.NetworkAlias.MarshalCBOR(w)
	}
	if n.NetworkIdentity != nil {
		return n.NetworkIdentity.MarshalCBOR(w)
	}
	if n.NetworkInfo != nil {
		return n.NetworkInfo.MarshalCBOR(w)
	}
	return fmt.Errorf("unable to marshal network identifier")
}

func (n *NetworkIdentifier) UnmarshalCBOR(r io.Reader) (err error) {
	doc := &DocumentBase{}
	err = doc.UnmarshalCBOR(r)
	if err != nil {
		return fmt.Errorf("unable to unmarshal network identifier into doc: %w", err)
	}

	switch doc.Type {
	case "core/network/alias":
		n.NetworkAlias = &NetworkAlias{}
		err := UnmarshalCBORBytes(doc.DocumentBytes, n.NetworkAlias)
		if err != nil {
			return fmt.Errorf("unable to unmarshal network alias: %w", err)
		}
	case "core/network":
		n.NetworkIdentity = &NetworkIdentity{}
		err := UnmarshalCBORBytes(doc.DocumentBytes, n.NetworkIdentity)
		if err != nil {
			return fmt.Errorf("unable to unmarshal network identity: %w", err)
		}
	case "core/network/info":
		n.NetworkInfo = &NetworkInfo{}
		err := UnmarshalCBORBytes(doc.DocumentBytes, n.NetworkInfo)
		if err != nil {
			return fmt.Errorf("unable to unmarshal network info: %w", err)
		}
	default:
		return fmt.Errorf("unknown network identifier type: %s", doc.Type)
	}
	return nil
}

func (n NetworkIdentifier) Hostname() string {
	if n.NetworkAlias != nil {
		return n.NetworkAlias.Hostname
	}
	if n.NetworkInfo != nil {
		return n.NetworkInfo.NetworkAlias.Hostname
	}
	return ""
}

func (n NetworkIdentifier) String() string {
	if n.NetworkAlias != nil {
		return n.NetworkAlias.String()
	}
	// TODO: implement NetworkIdentity.String() for NetworkIdentity
	// if n.NetworkIdentity != nil {
	// 	return n.NetworkIdentity.String()
	// }
	if n.NetworkInfo != nil {
		return n.NetworkInfo.String()
	}
	return ""
}
