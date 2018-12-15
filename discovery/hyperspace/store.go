package hyperspace

import (
	"sort"

	"github.com/james-bowman/sparse"

	"nimona.io/go/peers"
)

// NewStore retuns empty store
func NewStore() *Store {
	return &Store{
		peers: map[*sparse.Vector]*peers.PeerInfo{},
	}
}

// Store holds peer capabilities with their vectors
type Store struct {
	peers map[*sparse.Vector]*peers.PeerInfo
}

// Add peer capabilities to store
func (s *Store) Add(cs ...*peers.PeerInfo) {
	for _, c := range cs {
		v := Vectorise(getPeerInfoRequest(c))
		s.peers[v] = c
	}
}

// FindClosest returns peers that closest resemble the query
func (s *Store) FindClosest(q *peers.PeerInfoRequest) []*peers.PeerInfo {
	qv := Vectorise(q)

	// fmt.Println("-------- looking for", q)
	// fmt.Println("looking for", q)
	// fmt.Println("query vector", qv)

	type kv struct {
		distance  float64
		vector    *sparse.Vector
		peerInfos *peers.PeerInfo
	}

	r := []kv{}
	for v, c := range s.peers {
		d := CosineSimilarity(qv, v)
		// d := SimpleSimilarity(qv, v)
		r = append(r, kv{
			distance:  d,
			vector:    v,
			peerInfos: c,
		})
		// fmt.Println("--- distance from", v, "is", d)
	}

	sort.Slice(r, func(i, j int) bool {
		return r[i].distance > r[j].distance
	})

	// fmt.Println("first result is", r[0].peerInfos)
	// fmt.Println("first result vector is", r[0].vector)

	rs := []*peers.PeerInfo{}
	for i, c := range r {
		rs = append(rs, c.peerInfos)
		if i > 10 {
			break
		}
	}

	return rs
}
