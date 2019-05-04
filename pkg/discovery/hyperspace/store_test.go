package hyperspace

// import (
// 	"crypto/rand"
// 	"testing"

// 	"github.com/stretchr/testify/assert"

// 	"nimona.io/internal/encoding/base58"
// 	"nimona.io/pkg/crypto"
// 	"nimona.io/pkg/net/peer"
// )

// func TestStoreSimpleQuery(t *testing.T) {
// 	s := NewStore()

// 	cs := []*peer.PeerInfo{
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 		},
// 	}

// 	s.Add(cs...)

// 	for _, q := range cs {
// 		rs := s.FindClosest(getPeerInfoRequest(q))
// 		assert.Equal(t, q, rs[0])
// 	}
// }

// func TestStoreFindExact(t *testing.T) {
// 	s := NewStore()

// 	cs := []*peer.PeerInfo{
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 		},
// 	}

// 	s.Add(cs...)

// 	for _, c := range cs {
// 		q := &peer.PeerInfoRequest{
// 			AuthorityKeyHash: c.AuthorityKey.HashBase58(),
// 		}
// 		rs := s.FindClosest(q)
// 		assert.Equal(t, c.AuthorityKey.HashBase58(), rs[0].AuthorityKey.HashBase58())
// 	}
// }

// func TestStoreSimpleQueryWithNoise(t *testing.T) {
// 	s := NewStore()

// 	cs := []*peer.PeerInfo{
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 			Protocols: []string{
// 				"foo",
// 			},
// 			ContentTypes: []string{
// 				"bar",
// 			},
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 			Protocols: []string{
// 				"foo",
// 			},
// 			ContentTypes: []string{
// 				"bar",
// 			},
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 			Protocols: []string{
// 				"foo",
// 			},
// 			ContentTypes: []string{
// 				"bar",
// 			},
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 			Protocols: []string{
// 				"foo",
// 			},
// 			ContentTypes: []string{
// 				"bar",
// 			},
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 			Protocols: []string{
// 				"foo",
// 			},
// 			ContentTypes: []string{
// 				"bar",
// 			},
// 		},
// 	}

// 	s.Add(cs...)

// 	for _, c := range cs {
// 		q := &peer.PeerInfoRequest{
// 			AuthorityKeyHash: c.AuthorityKey.HashBase58(),
// 		}
// 		rs := s.FindClosest(q)
// 		assert.Equal(t, c.AuthorityKey.HashBase58(), rs[0].AuthorityKey.HashBase58())
// 	}
// }

// func TestStoreComplexQuery(t *testing.T) {
// 	s := NewStore()

// 	cs := []*peer.PeerInfo{
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 			Protocols: []string{
// 				"foo",
// 			},
// 			ContentTypes: []string{
// 				"bar",
// 			},
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 			Protocols: []string{
// 				"not-foo",
// 			},
// 			ContentTypes: []string{
// 				"not-bar",
// 			},
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 			Protocols: []string{
// 				"not-foo",
// 			},
// 			ContentTypes: []string{
// 				"bar",
// 			},
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 			Protocols: []string{
// 				"foo",
// 			},
// 			ContentTypes: []string{
// 				"not-bar",
// 			},
// 		},
// 		{
// 			AuthorityKey: getRandKey(),
// 			SignerKey:    getRandKey(),
// 			Protocols: []string{
// 				"very-not-foo",
// 			},
// 			ContentTypes: []string{
// 				"very-not-bar",
// 			},
// 		},
// 	}

// 	s.Add(cs...)

// 	for _, c := range cs {
// 		q := &peer.PeerInfoRequest{
// 			AuthorityKeyHash: c.AuthorityKey.HashBase58(),
// 		}
// 		rs := s.FindClosest(q)
// 		assert.Equal(t, c, rs[0])
// 	}

// 	for _, c := range cs {
// 		q := &peer.PeerInfoRequest{
// 			Protocols:    c.Protocols,
// 			ContentTypes: c.ContentTypes,
// 		}
// 		rs := s.FindClosest(q)
// 		assert.Equal(t, c, rs[0])
// 	}

// 	for _, c := range cs {
// 		q := &peer.PeerInfoRequest{
// 			SignerKeyHash: c.SignerKey.Fingerprint(),
// 			ContentTypes:  c.ContentTypes,
// 		}
// 		rs := s.FindClosest(q)
// 		assert.Equal(t, c, rs[0])
// 	}

// 	for _, c := range cs {
// 		q := &peer.PeerInfoRequest{
// 			AuthorityKeyHash: c.AuthorityKey.HashBase58(),
// 			ContentTypes:     c.ContentTypes,
// 		}
// 		rs := s.FindClosest(q)
// 		assert.Equal(t, c, rs[0])
// 	}

// 	// best effort

// 	for _, c := range cs {
// 		q := &peer.PeerInfoRequest{
// 			AuthorityKeyHash: c.AuthorityKey.HashBase58(),
// 			ContentTypes:     []string{"not here"},
// 		}
// 		rs := s.FindClosest(q)
// 		assert.Equal(t, c, rs[0])
// 	}

// 	for _, c := range cs {
// 		q := &peer.PeerInfoRequest{
// 			AuthorityKeyHash: c.AuthorityKey.HashBase58(),
// 			Protocols:        c.Protocols,
// 			ContentTypes:     []string{"not here"},
// 		}
// 		rs := s.FindClosest(q)
// 		assert.Equal(t, c, rs[0])
// 	}
// }

// func TestStoreSingleContentPerPeerQueryOne(t *testing.T) {
// 	s := NewStore()

// 	cs := []*peer.PeerInfo{}
// 	for i := 0; i < 100; i++ {
// 		cs = append(cs, &peer.PeerInfo{
// 			// AuthorityKey: getRandKey(),
// 			ContentTypes: []string{
// 				"foo",
// 			},
// 			ContentIDs: []string{
// 				base58.Encode(getRandBytes(32)),
// 				base58.Encode(getRandBytes(32)),
// 				base58.Encode(getRandBytes(32)),
// 				base58.Encode(getRandBytes(32)),
// 				base58.Encode(getRandBytes(32)),
// 				base58.Encode(getRandBytes(32)),
// 				base58.Encode(getRandBytes(32)),
// 				base58.Encode(getRandBytes(32)),
// 				base58.Encode(getRandBytes(32)),
// 			},
// 			SignerKey: getRandKey(),
// 		})
// 	}

// 	s.Add(cs...)

// 	for _, c := range cs[:10] {
// 		q := &peer.PeerInfoRequest{
// 			ContentIDs: []string{
// 				c.ContentIDs[0],
// 			},
// 		}
// 		rs := s.FindClosest(q)
// 		assert.Equal(t, q.ContentIDs[0], rs[0].ContentIDs[0])
// 	}
// }

// func TestStoreLowNumbers(t *testing.T) {
// 	testMultiplePeersMultipleContent(t, 100, 100, 10)
// }

// // func TestStoreManyPeers(t *testing.T) {
// // 	testMultiplePeersMultipleContent(t, 10000, 100, 10)
// // }

// // func TestStoreManyContentIDs(t *testing.T) {
// // 	testMultiplePeersMultipleContent(t, 100, 10000, 10)
// // }

// // func TestStoreManyPeersManyContentIDs(t *testing.T) {
// // 	if testing.Short() {
// // 		t.Skip("skipping annoyingly long 1000x1000 store test in short mode")
// // 	}
// // 	testMultiplePeersMultipleContent(t, 1000, 1000, 100)
// // }

// func testMultiplePeersMultipleContent(t *testing.T, nPeers, nContent, nCheck int) {
// 	if nCheck > nPeers {
// 		panic("cannot check more than what peers have")
// 	}

// 	s := NewStore()
// 	checkIDs := make([]string, nPeers)
// 	for i := 0; i < nPeers; i++ {
// 		cIDs := make([]string, nContent)
// 		for j := range cIDs {
// 			cIDs[j] = base58.Encode(getRandBytes(32))
// 		}
// 		checkIDs[i] = cIDs[0]
// 		c := &peer.PeerInfo{
// 			SignerKey: getRandKey(),
// 			// AuthorityKey: getRandKey(),
// 			ContentIDs: cIDs,
// 		}
// 		s.Add(c)
// 	}
// 	for _, cID := range checkIDs[:nCheck] {
// 		q := &peer.PeerInfoRequest{
// 			ContentIDs: []string{
// 				cID,
// 			},
// 		}
// 		rs := s.FindClosest(q)
// 		assert.Equal(t, q.ContentIDs[0], rs[0].ContentIDs[0])
// 	}
// }

// func getRandKey() *crypto.Key() {
// 	return &crypto.Key(){
// 		X: getRandBytes(32),
// 	}
// }

// func getRandBytes(n int) []byte {
// 	b := make([]byte, n)
// 	rand.Read(b)
// 	return b
// }
