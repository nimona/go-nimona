package nimona

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

func NewNetworkID(hostname string) NetworkID {
	return NetworkID{
		Hostname: hostname,
	}
}

type NetworkID struct {
	_        string `cborgen:"$prefix,const=nimona://net"`
	Hostname string
}

func (n NetworkID) String() string {
	return string(ResourceTypeNetworkHandle) + n.Hostname
}

func ParseNetworkID(nID string) (NetworkID, error) {
	prefix := string(ResourceTypeNetworkHandle)
	if !strings.HasPrefix(nID, prefix) {
		return NetworkID{}, fmt.Errorf("invalid resource id")
	}

	nID = strings.TrimPrefix(nID, prefix)
	return NetworkID{Hostname: nID}, nil
}

func (n NetworkID) Value() (driver.Value, error) {
	return n.String(), nil
}

func (n *NetworkID) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if netIDString, ok := value.(string); ok {
		netID, err := ParseNetworkID(netIDString)
		if err != nil {
			return fmt.Errorf("unable to scan into DocumentID: %w", err)
		}
		n.Hostname = netID.Hostname
		return nil
	}
	return fmt.Errorf("unable to scan %T into DocumentID", value)
}
