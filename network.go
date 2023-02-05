package nimona

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type (
	NetworkAlias struct {
		_        string `nimona:"$type,type=core/network/alias"`
		Hostname string `nimona:"hostname"`
	}
	NetworkIdentity struct {
		_                 string       `nimona:"$type,type=core/network/identity"`
		NetworkInfoRootID DocumentID   `nimona:"networkInfoRootID,omitempty"`
		NetworkAlias      NetworkAlias `nimona:"networkAlias,omitempty"`
	}
	NetworkInfo struct {
		Metadata      Metadata     `nimona:"$metadata,omitempty,type=core/network/info"`
		NetworkAlias  NetworkAlias `nimona:"networkAlias"`
		PeerAddresses []PeerAddr   `nimona:"peerAddresses"`
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

func (n *NetworkIdentity) NetworkIdentifier() NetworkIdentifier {
	return NetworkIdentifier{
		NetworkIdentity: n,
	}
}

func (n *NetworkInfo) NetworkIdentifier() NetworkIdentifier {
	return NetworkIdentifier{
		NetworkInfo: n,
	}
}

func (n *NetworkInfo) NetworkIdentity() NetworkIdentity {
	return NetworkIdentity{
		NetworkInfoRootID: NewDocumentID(n.DocumentMap()),
	}
}

func (n *NetworkInfo) String() string {
	return n.NetworkAlias.String()
}

func (n *NetworkIdentifier) DocumentMap() DocumentMap {
	if n.NetworkAlias != nil {
		return n.NetworkAlias.DocumentMap()
	}
	if n.NetworkIdentity != nil {
		return n.NetworkIdentity.DocumentMap()
	}
	if n.NetworkInfo != nil {
		return n.NetworkInfo.DocumentMap()
	}
	return DocumentMap{}
}

func (n *NetworkIdentifier) FromDocumentMap(m DocumentMap) {
	docType := m.Type()
	switch docType {
	case "core/network/alias":
		n.NetworkAlias = &NetworkAlias{}
		n.NetworkAlias.FromDocumentMap(m)
	case "core/network/identity":
		n.NetworkIdentity = &NetworkIdentity{}
		n.NetworkIdentity.FromDocumentMap(m)
	case "core/network/info":
		n.NetworkInfo = &NetworkInfo{}
		n.NetworkInfo.FromDocumentMap(m)
	}
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
