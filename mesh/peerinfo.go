package mesh

import (
	"fmt"
	"time"
)

type PeerInfo struct {
	ID        string
	Protocols map[string][]string
}

type peerInfoProtocol struct {
	PeerID      string
	Name        string
	Address     string
	LastUpdated time.Time
	Pinned      bool
}

// TODO maybe a better or just faster hash function?
func (p *peerInfoProtocol) Hash() string {
	return fmt.Sprintf("%s/%s/%s", p.PeerID, p.Name, p.Address)
}
