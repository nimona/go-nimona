package nimona

import (
	"fmt"
	"strings"
)

type NetworkID struct {
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
