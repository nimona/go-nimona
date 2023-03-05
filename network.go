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
		NetworkInfoRootID: NewDocumentID(n.Document()),
	}
}

func (n *NetworkInfo) String() string {
	return n.NetworkAlias.String()
}

func (n *NetworkIdentifier) Document() *Document {
	if n.NetworkAlias != nil {
		return n.NetworkAlias.Document()
	}
	if n.NetworkIdentity != nil {
		return n.NetworkIdentity.Document()
	}
	if n.NetworkInfo != nil {
		return n.NetworkInfo.Document()
	}
	return nil
}

func (n *NetworkIdentifier) FromDocument(m *Document) error {
	docType := m.Type()
	switch docType {
	case "core/network/alias":
		n.NetworkAlias = &NetworkAlias{}
		err := n.NetworkAlias.FromDocument(m)
		if err != nil {
			return fmt.Errorf("unable to parse network alias: %w", err)
		}
		return nil
	case "core/network/identity":
		n.NetworkIdentity = &NetworkIdentity{}
		err := n.NetworkIdentity.FromDocument(m)
		if err != nil {
			return fmt.Errorf("unable to parse network identity: %w", err)
		}
		return nil
	case "core/network/info":
		n.NetworkInfo = &NetworkInfo{}
		err := n.NetworkInfo.FromDocument(m)
		if err != nil {
			return fmt.Errorf("unable to parse network info: %w", err)
		}
		return nil
	}

	return fmt.Errorf("unable to parse network identifier from document map, unknown type %s", docType)
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
