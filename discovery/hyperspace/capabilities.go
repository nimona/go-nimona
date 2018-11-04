package hyperspace

import (
	"fmt"

	"github.com/james-bowman/sparse"
)

// PeerCapabilities describes a peer's capabilities
type PeerCapabilities struct {
	IdentityKey string
	PeerKey     string
	Protocols   []string
	Resources   []string
	ContentIDs  []string
}

// Vectorise returns a sparse vector from PeerCapabilities
func Vectorise(c *PeerCapabilities) *sparse.Vector {
	i := []int{}
	b := Hash([]byte("identitykey_" + c.IdentityKey))
	i = append(i, int(b))
	b = Hash([]byte("peerkey_" + c.PeerKey))
	i = append(i, int(b))
	for _, protocol := range c.Protocols {
		b = Hash([]byte("protocol_" + protocol))
		i = append(i, int(b))
	}
	for _, resouce := range c.Resources {
		b = Hash([]byte("resource_" + resouce))
		i = append(i, int(b))
	}
	for _, contentid := range c.ContentIDs {
		for _, c := range chunk([]byte(contentid), 1) {
			b = Hash([]byte(fmt.Sprintf("c_%d_%s", i, string(c))))
			i = append(i, int(b))
		}
	}
	d := []float64{}
	for range i {
		d = append(d, 1)
	}
	return sparse.NewVector(int(scaledMax), i, d)
}
