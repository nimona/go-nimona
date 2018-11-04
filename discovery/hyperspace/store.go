package hyperspace

import (
	"fmt"
	"sort"

	"github.com/james-bowman/sparse"
)

// NewStore retuns empty store
func NewStore() *Store {
	return &Store{
		peers: map[*sparse.Vector]*PeerCapabilities{},
	}
}

// Store holds peer capabilities with their vectors
type Store struct {
	peers map[*sparse.Vector]*PeerCapabilities
}

// Add peer capabilities to store
func (s *Store) Add(cs ...*PeerCapabilities) {
	for _, c := range cs {
		v := Vectorise(c)
		s.peers[v] = c
		fmt.Println("Adding", c, "as", v)
	}
}

// FindClosest returns peers that closest resemble the query
func (s *Store) FindClosest(q *PeerCapabilities) []*PeerCapabilities {
	qv := Vectorise(q)

	type kv struct {
		distance         float64
		vector           *sparse.Vector
		peerCapabilities *PeerCapabilities
	}

	r := []kv{}
	for v, c := range s.peers {
		d := CosineSimilarity(qv, v)
		r = append(r, kv{
			distance:         d,
			vector:           v,
			peerCapabilities: c,
		})
	}

	sort.Slice(r, func(i, j int) bool {
		return r[i].distance > r[j].distance
	})

	rs := []*PeerCapabilities{}
	for i, c := range r {
		rs = append(rs, c.peerCapabilities)
		if i > 10 {
			break
		}
	}

	return rs
}
