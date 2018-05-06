package mesh

import (
	"fmt"
	"time"
)

type PeerInfo struct {
	ID        string              `json:"id"`
	Protocols map[string][]string `json:"protocols"`
}

type peerInfoProtocol struct {
	PeerID      string    `json:"peer_id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	LastUpdated time.Time `json:"last_updated,omitempty"`
	Pinned      bool      `json:"pinned,omitempty"`
}

// TODO maybe a better or just faster hash function?
func (p *peerInfoProtocol) Hash() string {
	return fmt.Sprintf("%s/%s/%s", p.PeerID, p.Name, p.Address)
}
