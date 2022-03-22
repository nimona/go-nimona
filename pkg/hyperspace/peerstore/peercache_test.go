package peerstore

import (
	"testing"
	"time"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/tilde"

	"github.com/stretchr/testify/assert"
)

func TestPeerCache_Lookup(t *testing.T) {
	opk, err := crypto.NewEd25519PrivateKey()
	assert.NoError(t, err)

	opk2, err := crypto.NewEd25519PrivateKey()
	assert.NoError(t, err)

	p1 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: opk.PublicKey().DID(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Metadata: object.Metadata{
				Owner: opk.PublicKey().DID(),
			},
		},
		Digests: []tilde.Digest{"foo", "bar"},
	}

	p2 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: opk2.PublicKey().DID(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Metadata: object.Metadata{
				Owner: opk2.PublicKey().DID(),
			},
		},
		Digests: []tilde.Digest{"foo", "not-bar"},
	}

	pc := NewPeerCache(200*time.Millisecond, "test0")

	pc.Put(p1, 200*time.Millisecond)
	pc.Put(p2, 200*time.Millisecond)

	ps := pc.LookupByDigest("foo")
	assert.ElementsMatch(t, []*hyperspace.Announcement{p1, p2}, ps)

	ps = pc.LookupByDigest("bar")
	assert.ElementsMatch(t, []*hyperspace.Announcement{p1}, ps)

	ps = pc.LookupByDigest("not-bar")
	assert.ElementsMatch(t, []*hyperspace.Announcement{p2}, ps)
}

func TestPeerCache_List(t *testing.T) {
	opk, err := crypto.NewEd25519PrivateKey()
	assert.NoError(t, err)

	opk2, err := crypto.NewEd25519PrivateKey()
	assert.NoError(t, err)

	p1a := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: opk.PublicKey().DID(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Metadata: object.Metadata{
				Owner: opk.PublicKey().DID(),
			},
			Addresses: []string{"foo"},
		},
	}

	p1b := &hyperspace.Announcement{
		Version: 1,
		Metadata: object.Metadata{
			Owner: opk.PublicKey().DID(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Metadata: object.Metadata{
				Owner: opk.PublicKey().DID(),
			},
			Addresses: []string{"bar"},
		},
	}

	p2 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: opk2.PublicKey().DID(),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Metadata: object.Metadata{
				Owner: opk2.PublicKey().DID(),
			},
			Addresses: []string{"foo"},
		},
	}

	pc := NewPeerCache(200*time.Millisecond, "test1")

	ok := pc.Put(p1a, 200*time.Millisecond)
	assert.True(t, ok)
	ok = pc.Put(p1a, 200*time.Millisecond)
	assert.False(t, ok)
	ok = pc.Put(p1b, 200*time.Millisecond)
	assert.True(t, ok)
	ok = pc.Put(p2, 200*time.Millisecond)
	assert.True(t, ok)

	ps := pc.List()
	assert.ElementsMatch(t, []*hyperspace.Announcement{p1b, p2}, ps)
}

func TestPeerCache_Remove(t *testing.T) {
	opk, err := crypto.NewEd25519PrivateKey()
	assert.NoError(t, err)

	pc := NewPeerCache(200*time.Millisecond, "test2")

	pc.Put(
		&hyperspace.Announcement{
			Metadata: object.Metadata{
				Owner: opk.PublicKey().DID(),
			},
			ConnectionInfo: &peer.ConnectionInfo{
				Metadata: object.Metadata{
					Owner: opk.PublicKey().DID(),
				},
			},
		},
		200*time.Millisecond,
	)

	pc.Remove(opk.PublicKey().DID())

	pr, err := pc.Get(opk.PublicKey().DID())
	assert.Error(t, err)
	assert.Nil(t, pr)
}

func TestPeerCache_Touch(t *testing.T) {
	opk, err := crypto.NewEd25519PrivateKey()
	assert.NoError(t, err)

	pc := NewPeerCache(200*time.Millisecond, "test3")

	pc.Put(
		&hyperspace.Announcement{
			Metadata: object.Metadata{
				Owner: opk.PublicKey().DID(),
			},
			ConnectionInfo: &peer.ConnectionInfo{
				Metadata: object.Metadata{
					Owner: opk.PublicKey().DID(),
				},
			},
		},
		200*time.Millisecond,
	)

	time.Sleep(100 * time.Millisecond)

	pc.Touch(opk.PublicKey().DID(), 300*time.Millisecond)

	time.Sleep(200 * time.Millisecond)

	pr, err := pc.Get(opk.PublicKey().DID())
	assert.NoError(t, err)
	assert.True(t, opk.PublicKey().DID().Equals(
		pr.ConnectionInfo.Metadata.Owner,
	))

	time.Sleep(300 * time.Millisecond)

	pr, err = pc.Get(opk.PublicKey().DID())
	assert.Error(t, err)
	assert.Nil(t, pr)

	pc.Touch(opk.PublicKey().DID(), 300*time.Millisecond)
}

func TestPeerCache_TTL(t *testing.T) {
	opk, err := crypto.NewEd25519PrivateKey()
	assert.NoError(t, err)

	pc := NewPeerCache(200*time.Millisecond, "test4")

	pc.Put(
		&hyperspace.Announcement{
			Metadata: object.Metadata{
				Owner: opk.PublicKey().DID(),
			},
			ConnectionInfo: &peer.ConnectionInfo{
				Metadata: object.Metadata{
					Owner: opk.PublicKey().DID(),
				},
			},
		},
		600*time.Millisecond,
	)

	pr, err := pc.Get(opk.PublicKey().DID())
	assert.NoError(t, err)
	assert.True(t, opk.PublicKey().DID().Equals(
		pr.ConnectionInfo.Metadata.Owner,
	))

	time.Sleep(900 * time.Millisecond)

	pr, err = pc.Get(opk.PublicKey().DID())
	assert.Error(t, err)
	assert.Nil(t, pr)
}
