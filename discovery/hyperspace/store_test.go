package hyperspace

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestStoreSimpleQuery(t *testing.T) {
	s := NewStore()

	cs := []*PeerCapabilities{
		&PeerCapabilities{
			IdentityKey: "1",
		},
		&PeerCapabilities{
			IdentityKey: "2",
		},
		&PeerCapabilities{
			IdentityKey: "3",
		},
		&PeerCapabilities{
			IdentityKey: "4",
		},
		&PeerCapabilities{
			IdentityKey: "5",
		},
	}

	s.Add(cs...)

	for _, q := range cs {
		rs := s.FindClosest(q)
		assert.Equal(t, q, rs[0])
	}
}

func TestStoreSimpleQueryWithNoise(t *testing.T) {
	s := NewStore()

	cs := []*PeerCapabilities{
		&PeerCapabilities{
			IdentityKey: "1",
			PeerKey:     "A",
			Protocols: []string{
				"foo",
			},
			Resources: []string{
				"bar",
			},
		},
		&PeerCapabilities{
			IdentityKey: "2",
			PeerKey:     "A",
			Protocols: []string{
				"foo",
			},
			Resources: []string{
				"bar",
			},
		},
		&PeerCapabilities{
			IdentityKey: "3",
			PeerKey:     "A",
			Protocols: []string{
				"foo",
			},
			Resources: []string{
				"bar",
			},
		},
		&PeerCapabilities{
			IdentityKey: "4",
			PeerKey:     "A",
			Protocols: []string{
				"foo",
			},
			Resources: []string{
				"bar",
			},
		},
		&PeerCapabilities{
			IdentityKey: "5",
			PeerKey:     "A",
			Protocols: []string{
				"foo",
			},
			Resources: []string{
				"bar",
			},
		},
	}

	s.Add(cs...)

	for _, c := range cs {
		q := &PeerCapabilities{
			IdentityKey: c.IdentityKey,
		}
		rs := s.FindClosest(q)
		assert.Equal(t, c, rs[0])
	}
}

func TestStoreComplexQuery(t *testing.T) {
	s := NewStore()

	cs := []*PeerCapabilities{
		&PeerCapabilities{
			IdentityKey: "1",
			PeerKey:     "A",
			Protocols: []string{
				"foo",
			},
			Resources: []string{
				"bar",
			},
		},
		&PeerCapabilities{
			IdentityKey: "2",
			PeerKey:     "A",
			Protocols: []string{
				"not-foo",
			},
			Resources: []string{
				"not-bar",
			},
		},
		&PeerCapabilities{
			IdentityKey: "3",
			PeerKey:     "B",
			Protocols: []string{
				"not-foo",
			},
			Resources: []string{
				"bar",
			},
		},
		&PeerCapabilities{
			IdentityKey: "4",
			PeerKey:     "B",
			Protocols: []string{
				"foo",
			},
			Resources: []string{
				"not-bar",
			},
		},
		&PeerCapabilities{
			IdentityKey: "5",
			PeerKey:     "B",
			Protocols: []string{
				"very-not-foo",
			},
			Resources: []string{
				"very-not-bar",
			},
		},
	}

	s.Add(cs...)

	for _, c := range cs {
		q := &PeerCapabilities{
			IdentityKey: c.IdentityKey,
		}
		rs := s.FindClosest(q)
		assert.Equal(t, c, rs[0])
	}

	for _, c := range cs {
		q := &PeerCapabilities{
			Protocols: c.Protocols,
			Resources: c.Resources,
		}
		rs := s.FindClosest(q)
		assert.Equal(t, c, rs[0])
	}

	for _, c := range cs {
		q := &PeerCapabilities{
			PeerKey:   c.PeerKey,
			Resources: c.Resources,
		}
		rs := s.FindClosest(q)
		assert.Equal(t, c, rs[0])
	}

	for _, c := range cs {
		q := &PeerCapabilities{
			IdentityKey: c.IdentityKey,
			Resources:   c.Resources,
		}
		rs := s.FindClosest(q)
		assert.Equal(t, c, rs[0])
	}

	// best effort

	for _, c := range cs {
		q := &PeerCapabilities{
			IdentityKey: c.IdentityKey,
			Resources:   []string{"not here"},
		}
		rs := s.FindClosest(q)
		assert.Equal(t, c, rs[0])
	}

	for _, c := range cs {
		q := &PeerCapabilities{
			IdentityKey: c.IdentityKey,
			Protocols:   c.Protocols,
			Resources:   []string{"not here"},
		}
		rs := s.FindClosest(q)
		assert.Equal(t, c, rs[0])
	}
}

func TestStoreSingleContentPerPeerQueryOne(t *testing.T) {
	s := NewStore()

	cs := []*PeerCapabilities{}
	for i := 0; i < 100; i++ {
		cs = append(cs, &PeerCapabilities{
			// IdentityKey: uuid.NewV4().String(),
			Resources: []string{
				"foo",
			},
			ContentIDs: []string{
				uuid.NewV4().String(),
				uuid.NewV4().String(),
				uuid.NewV4().String(),
				uuid.NewV4().String(),
				uuid.NewV4().String(),
				uuid.NewV4().String(),
				uuid.NewV4().String(),
				uuid.NewV4().String(),
				uuid.NewV4().String(),
			},
		})
	}

	s.Add(cs...)

	for _, c := range cs {
		q := &PeerCapabilities{
			ContentIDs: []string{
				c.ContentIDs[0],
			},
		}
		rs := s.FindClosest(q)
		assert.Equal(t, q.ContentIDs[0], rs[0].ContentIDs[0])
	}
}

func TestStoreMultipleContentsPerPeerQueryOne(t *testing.T) {
	s := NewStore()

	cs := []*PeerCapabilities{}
	for i := 0; i < 1000; i++ {
		cs = append(cs, &PeerCapabilities{
			IdentityKey: uuid.NewV4().String(),
			ContentIDs: []string{
				uuid.NewV4().String(),
				uuid.NewV4().String(),
				uuid.NewV4().String(),
				uuid.NewV4().String(),
				uuid.NewV4().String(),
				uuid.NewV4().String(),
				uuid.NewV4().String(),
				uuid.NewV4().String(),
				uuid.NewV4().String(),
			},
		})
	}

	s.Add(cs...)

	for _, q := range cs[:100] {
		rs := s.FindClosest(q)
		assert.Equal(t, q, rs[0])
	}
}
