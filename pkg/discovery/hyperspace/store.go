package hyperspace

import (
	"sort"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/discovery/hyperspace/bloom"
	"nimona.io/pkg/peer"
)

//go:generate $GOBIN/genny -in=$GENERATORS/syncmap/syncmap.go -out=syncmap_fingerprint_bloom_generated.go -pkg hyperspace gen "KeyType=crypto.Fingerprint ValueType=Announced"
//go:generate $GOBIN/genny -in=$GENERATORS/syncmap/syncmap.go -out=syncmap_fingerprint_peer_generated.go -pkg hyperspace gen "KeyType=crypto.Fingerprint ValueType=peer.Peer"

// NewStore retuns empty store
func NewStore() *Store {
	return &Store{
		blooms: &CryptoFingerprintAnnouncedSyncMap{},
		peers:  &CryptoFingerprintPeerPeerSyncMap{},
	}
}

// Store holds peer content blooms and their fingerprints
type Store struct {
	blooms *CryptoFingerprintAnnouncedSyncMap
	peers  *CryptoFingerprintPeerPeerSyncMap
}

// Add peers
func (s *Store) AddPeer(peer *peer.Peer) {
	s.peers.Put(peer.Signature.PublicKey.Fingerprint(), peer)
}

// Add content hashes
func (s *Store) AddContentHashes(c *Announced) {
	s.blooms.Put(c.Signature.PublicKey.Fingerprint(), c)
}

// FindClosestPeer returns peers that closest resemble the query
func (s *Store) FindClosestPeer(f crypto.Fingerprint) []*peer.Peer {
	q := bloom.NewBloom(f.String())

	type kv struct {
		bloomIntersection int
		peer              *peer.Peer
	}

	r := []kv{}
	s.peers.Range(func(f crypto.Fingerprint, p *peer.Peer) bool {
		pb := bloom.NewBloom(p.Signature.PublicKey.Fingerprint().String())
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
func (s *Store) FindClosestContentProvider(q bloom.Bloomer) []*Announced {
	type kv struct {
		bloomIntersection int
		contentHashBloom  *Announced
	}

	r := []kv{}
	s.blooms.Range(func(f crypto.Fingerprint, b *Announced) bool {
		r = append(r, kv{
			bloomIntersection: intersectionCount(q.Bloom(), b.Bloom()),
			contentHashBloom:  b,
		})
		return true
	})

	sort.Slice(r, func(i, j int) bool {
		return r[i].bloomIntersection < r[j].bloomIntersection
	})

	cs := []*Announced{}
	for i, c := range r {
		cs = append(cs, c.contentHashBloom)
		if i > 10 { // TODO make limit configurable
			break
		}
	}

	return cs
}

// FindByFingerprint returns peers that are signed by a fingerprint
func (s *Store) FindByFingerprint(
	fingerprint crypto.Fingerprint,
) []*peer.Peer {
	ps := []*peer.Peer{}
	s.peers.Range(func(f crypto.Fingerprint, p *peer.Peer) bool {
		if peerMatchesKeyFingerprint(p, fingerprint) {
			ps = append(ps, p)
		}
		return true
	})
	return ps
}

func peerMatchesKeyFingerprint(
	peer *peer.Peer,
	fingerprint crypto.Fingerprint,
) bool {
	for _, k := range crypto.GetSignatureKeys(peer.Signature) {
		if k.Fingerprint().String() == fingerprint.String() {
			return true
		}
	}
	return false
}

// FindByContent returns peers that match a given content hash
func (s *Store) FindByContent(b bloom.Bloomer) []*Announced {
	cs := []*Announced{}
	q := b.Bloom()

	s.blooms.Range(func(f crypto.Fingerprint, c *Announced) bool {
		if intersectionCount(c.Bloom(), b.Bloom()) != len(q) {
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
