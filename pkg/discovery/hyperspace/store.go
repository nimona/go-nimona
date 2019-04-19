package hyperspace

import (
	"sort"
	"sync"

	"github.com/james-bowman/sparse"

	"nimona.io/pkg/net/peer"
)

type storeValue struct {
	vector   *sparse.Vector
	peerInfo *peer.PeerInfo
}

// NewStore retuns empty store
func NewStore() *Store {
	return &Store{
		peers: map[string]*storeValue{},
	}
}

// Store holds peer capabilities with their vectors
type Store struct {
	lock  sync.RWMutex
	peers map[string]*storeValue
}

// Add peer capabilities to store
func (s *Store) Add(cs ...*peer.PeerInfo) {
	s.lock.Lock()
	for _, c := range cs {
		v := Vectorise(getPeerInfoRequest(c))
		s.peers[c.SignerKey.HashBase58()] = &storeValue{
			vector:   v,
			peerInfo: c,
		}
	}
	s.lock.Unlock()
}

// FindClosest returns peers that closest resemble the query
func (s *Store) FindClosest(q *peer.PeerInfoRequest) []*peer.PeerInfo {
	s.lock.RLock()
	defer s.lock.RUnlock()

	qv := Vectorise(q)

	// fmt.Println("-------- looking for", q)
	// fmt.Println("looking for", q)
	// fmt.Println("query vector", qv)

	type kv struct {
		distance  float64
		vector    *sparse.Vector
		peerInfos *peer.PeerInfo
	}

	r := []kv{}
	for _, v := range s.peers {
		d := CosineSimilarity(qv, v.vector)
		// d := SimpleSimilarity(qv, v)
		r = append(r, kv{
			distance:  d,
			vector:    v.vector,
			peerInfos: v.peerInfo,
		})
		// fmt.Println("--- distance from", v, "is", d)
	}

	sort.Slice(r, func(i, j int) bool {
		return r[i].distance > r[j].distance
	})

	// fmt.Println("first result is", r[0].peerInfos)
	// fmt.Println("first result vector is", r[0].vector)

	rs := []*peer.PeerInfo{}
	for i, c := range r {
		rs = append(rs, c.peerInfos)
		if i > 10 {
			break
		}
	}

	return rs
}

// FindExact returns peers that match query authority or peer
func (s *Store) FindExact(q *peer.PeerInfoRequest) []*peer.PeerInfo {
	s.lock.RLock()
	defer s.lock.RUnlock()

	ps := []*peer.PeerInfo{}
check:
	for _, v := range s.peers {
		p := v.peerInfo
		if q.AuthorityKeyHash != "" && p.AuthorityKey != nil &&
			q.AuthorityKeyHash == p.AuthorityKey.HashBase58() {
			ps = append(ps, p)
			continue
		}
		if q.SignerKeyHash != "" && p.SignerKey != nil &&
			q.SignerKeyHash == p.SignerKey.HashBase58() {
			ps = append(ps, p)
			continue
		}
		for _, ch := range p.ContentIDs {
			for _, rch := range q.ContentIDs {
				if ch == rch {
					ps = append(ps, p)
					break check
				}
			}
		}
	}
	return ps
}
