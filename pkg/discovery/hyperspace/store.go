package hyperspace

import (
	"sort"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/bloom"
	"nimona.io/pkg/peer"
)

//go:generate $GOBIN/genny -in=$GENERATORS/syncmap/syncmap.go -out=syncmap_publickey_peer_generated.go -pkg hyperspace gen "KeyType=crypto.PublicKey ValueType=peer.Peer"

// NewStore retuns empty store
func NewStore() *Store {
	return &Store{
		peers: &CryptoPublicKeyPeerPeerSyncMap{},
	}
}

// Store holds peer content blooms and their fingerprints
type Store struct {
	peers *CryptoPublicKeyPeerPeerSyncMap
}

// Add peers
func (s *Store) AddPeer(p *peer.Peer) {
	xp, ok := s.peers.Get(p.Header.Signature.Signer)
	if ok && xp != nil && xp.Version != 0 && xp.Version >= p.Version {
		return
	}
	s.peers.Put(p.Header.Signature.Signer, p)
}

// GetClosest returns peers that closest resemble the query
func (s *Store) GetClosest(q bloom.Bloom) []*peer.Peer {
	type kv struct {
		bloomIntersection int
		peer              *peer.Peer
	}

	r := []kv{}
	s.peers.Range(func(f crypto.PublicKey, p *peer.Peer) bool {
		r = append(r, kv{
			bloomIntersection: intersectionCount(
				q.Bloom(),
				p.Bloom,
			),
			peer: p,
		})
		return true
	})

	sort.Slice(r, func(i, j int) bool {
		return r[i].bloomIntersection < r[j].bloomIntersection
	})

	fs := []*peer.Peer{}
	for i, c := range r {
		fs = append(fs, c.peer)
		if i > 10 { // TODO make limit configurable
			break
		}
	}

	return fs
}

// Get returns peers that match a query
func (s *Store) Get(q bloom.Bloom) []*peer.Peer {
	ps := []*peer.Peer{}
	s.peers.Range(func(f crypto.PublicKey, p *peer.Peer) bool {
		if bloom.Bloom(p.Bloom).Contains(q) {
			ps = append(ps, p)
		}
		return true
	})
	return ps
}

func intersectionCount(a, b []int64) int {
	m := make(map[int64]uint64)
	for _, k := range a {
		m[k] |= (1 << 0)
	}
	for _, k := range b {
		m[k] |= (1 << 1)
	}

	i := 0
	for _, v := range m {
		a := v&(1<<0) != 0
		b := v&(1<<1) != 0
		switch {
		case a && b:
			i++
		}
	}

	return i
}
