package hyperspace

import (
	"sort"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery/hyperspace/bloom"
	"nimona.io/pkg/peer"
)

//go:generate $GOBIN/genny -in=$GENERATORS/syncmap/syncmap.go -out=syncmap_publickey_bloom_generated.go -pkg hyperspace gen "KeyType=crypto.PublicKey ValueType=peer.Peer"
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
	s.peers.Range(func(k crypto.PublicKey, v *peer.Peer) bool {
		return true
	})
	s.peers.Put(p.Signature.Signer, p)
}

// FindClosestPeer returns peers that closest resemble the query
func (s *Store) FindClosestPeer(f crypto.PublicKey) []*peer.Peer {
	q := bloom.New(f.String())

	type kv struct {
		bloomIntersection int
		peer              *peer.Peer
	}

	r := []kv{}
	s.peers.Range(func(f crypto.PublicKey, p *peer.Peer) bool {
		pb := bloom.New(p.Signature.Signer.String())
		r = append(r, kv{
			bloomIntersection: intersectionCount(q.Bloom(), pb.Bloom()),
			peer:              p,
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

// FindClosestContentProvider returns peers that are "closet" to the given
// content
func (s *Store) FindClosestContentProvider(q []int64) []*peer.Peer {
	type kv struct {
		bloomIntersection int
		peer              *peer.Peer
	}

	r := []kv{}
	s.peers.Range(func(f crypto.PublicKey, p *peer.Peer) bool {
		r = append(r, kv{
			bloomIntersection: intersectionCount(q, p.ContentBloom),
			peer:              p,
		})
		return true
	})

	sort.Slice(r, func(i, j int) bool {
		return r[i].bloomIntersection < r[j].bloomIntersection
	})

	cs := []*peer.Peer{}
	for i, c := range r {
		cs = append(cs, c.peer)
		if i > 10 { // TODO make limit configurable
			break
		}
	}

	return cs
}

// FindByPublicKey returns peers that are signed by a fingerprint
func (s *Store) FindByPublicKey(
	publicKey crypto.PublicKey,
) []*peer.Peer {
	ps := []*peer.Peer{}
	s.peers.Range(func(f crypto.PublicKey, p *peer.Peer) bool {
		for _, cert := range p.Certificates {
			if cert.Signature.Signer.Equals(publicKey) {
				ps = append(ps, p)
			}
		}
		return true
	})
	return ps
}

// FindByContent returns peers that match a given content hash
func (s *Store) FindByContent(q []int64) []*peer.Peer {
	cs := []*peer.Peer{}

	s.peers.Range(func(f crypto.PublicKey, c *peer.Peer) bool {
		if intersectionCount(c.ContentBloom, q) != len(q) {
			return true
		}
		cs = append(cs, c)
		return true
	})

	return cs
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
