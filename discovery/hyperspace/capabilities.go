package hyperspace

import (
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
	if c.IdentityKey != "" {
		i = append(i, HashChunked("ik", []byte(c.IdentityKey))...)
	}
	if c.PeerKey != "" {
		i = append(i, HashChunked("pk", []byte(c.PeerKey))...)
	}
	if len(c.Protocols) > 0 {
		for _, protocol := range c.Protocols {
			i = append(i, HashChunked("p", []byte(protocol))...)
		}
	}
	if len(c.Resources) > 0 {
		for _, resource := range c.Resources {
			i = append(i, HashChunked("r", []byte(resource))...)
		}
	}
	if len(c.ContentIDs) > 0 {
		for _, contentid := range c.ContentIDs {
			i = append(i, HashChunked("cid", []byte(contentid))...)
		}
	}
	d := []float64{}
	for range i {
		d = append(d, 1)
	}
	v := sparse.NewVector(int(scaledMax), i, d)
	return v
}
