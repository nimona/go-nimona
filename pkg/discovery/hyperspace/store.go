package hyperspace

import (
	"fmt"
	"sort"

	"nimona.io/pkg/discovery/hyperspace/bloom"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/peer"
)

//go:generate $GOBIN/genny -in=../../../internal/generator/syncmap/syncmap.go -out=syncmap_fingerprint_bloom_generated.go -pkg hyperspace gen "KeyType=crypto.Fingerprint ValueType=ContentHashesBloom"
//go:generate $GOBIN/genny -in=../../../internal/generator/syncmap/syncmap.go -out=syncmap_fingerprint_peer_generated.go -pkg hyperspace gen "KeyType=crypto.Fingerprint ValueType=peer.Peer"

// NewStore retuns empty store
func NewStore() *Store {
	return &Store{
		blooms: &CryptoFingerprintContentHashesBloomSyncMap{},
		peers:  &CryptoFingerprintPeerPeerSyncMap{},
	}
}

// Store holds peer content blooms and their fingerprints
type Store struct {
	blooms *CryptoFingerprintContentHashesBloomSyncMap
	peers  *CryptoFingerprintPeerPeerSyncMap
}

// Add peers
func (s *Store) AddPeer(peer *peer.Peer) {
	s.peers.Put(peer.Fingerprint(), peer)
}

// Add content hashes
func (s *Store) AddContentHashes(c *ContentHashesBloom) {
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
		pb := bloom.NewBloom(p.Fingerprint().String())
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
func (s *Store) FindClosestContentProvider(q bloom.Bloomer) []*ContentHashesBloom {
	type kv struct {
		bloomIntersection int
		contentHashBloom  *ContentHashesBloom
	}

	r := []kv{}
	s.blooms.Range(func(f crypto.Fingerprint, b *ContentHashesBloom) bool {
		r = append(r, kv{
			bloomIntersection: intersectionCount(q.Bloom(), b.Bloom()),
			contentHashBloom:  b,
		})
		return true
	})

	sort.Slice(r, func(i, j int) bool {
		return r[i].bloomIntersection < r[j].bloomIntersection
	})

	cs := []*ContentHashesBloom{}
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
		for _, k := range crypto.GetSignatureKeys(p.Signature) {
			fmt.Println("_____", k.Fingerprint().String(), fingerprint.String())
			if k.Fingerprint().String() != fingerprint.String() {
				continue
			}
			ps = append(ps, p)
			break
		}
		return true
	})
	return ps
}

// FindByContent returns peers that match a given content hash
func (s *Store) FindByContent(contentHash string) []*ContentHashesBloom {
	cs := []*ContentHashesBloom{}
	q := bloom.NewBloom(contentHash)

	s.blooms.Range(func(f crypto.Fingerprint, c *ContentHashesBloom) bool {
		if intersectionCount(c.Bloom(), q.Bloom()) != len(q) {
			return true
		}
		cs = append(cs, c)
		return true
	})

	return cs
}

func intersectionCount(a, b []int) int {
	m := make(map[int]uint64)
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
