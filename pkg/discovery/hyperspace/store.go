package hyperspace

import (
	"sort"
	"sync"

	"github.com/james-bowman/sparse"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/net/peer"
)

type storeValue struct {
	vector   *sparse.Vector
	peerInfo *peer.PeerInfo
}

// NewStore retuns empty store
func NewStore() *Store {
	return &Store{
		peers: map[crypto.Fingerprint]*storeValue{},
	}
}

// Store holds peer capabilities with their vectors
type Store struct {
	lock  sync.RWMutex
	peers map[crypto.Fingerprint]*storeValue
}

// Add peer capabilities to store
func (s *Store) Add(cs ...*peer.PeerInfo) {
	s.lock.Lock()
	for _, c := range cs {
		pir := getPeerInfoRequest(c)
		v := Vectorise(pir)
		fingerprint := c.Fingerprint()
		s.peers[fingerprint] = &storeValue{
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
		// fmt.Println("--- distance from", v.peerInfo.Fingerprint(), "is", d)
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

// FindByFingerprint returns peers that are signed by a fingerprint
func (s *Store) FindByFingerprint(fingerprint crypto.Fingerprint) []*peer.PeerInfo {
	s.lock.RLock()
	defer s.lock.RUnlock()

	ps := map[crypto.Fingerprint]*peer.PeerInfo{}
	for _, v := range s.peers {
		p := v.peerInfo
		keys := crypto.GetSignatureKeys(p.Signature)
		for _, k := range keys {
			if k.Fingerprint() == fingerprint {
				ps[fingerprint] = p
				break
			}
		}
	}

	fps := []*peer.PeerInfo{}
	for _, p := range ps {
		fps = append(fps, p)
	}

	return fps
}

// FindByContent returns peers that match a given content hash
func (s *Store) FindByContent(contentHash string) []*peer.PeerInfo {
	s.lock.RLock()
	defer s.lock.RUnlock()

	ps := map[crypto.Fingerprint]*peer.PeerInfo{}
	for _, v := range s.peers {
		p := v.peerInfo
		for _, ch := range p.ContentIDs {
			if ch == contentHash {
				ps[p.Fingerprint()] = p
				break
			}
		}
	}

	fps := []*peer.PeerInfo{}
	for _, p := range ps {
		fps = append(fps, p)
	}

	return fps
}
