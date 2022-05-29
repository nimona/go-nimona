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
			Owner: peer.IDFromPublicKey(opk.PublicKey()),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Owner: peer.IDFromPublicKey(opk.PublicKey()),
		},
		Digests: []tilde.Digest{"foo", "bar"},
	}

	p2 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: peer.IDFromPublicKey(opk2.PublicKey()),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Owner: peer.IDFromPublicKey(opk2.PublicKey()),
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
			Owner: peer.IDFromPublicKey(opk.PublicKey()),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Owner:     peer.IDFromPublicKey(opk.PublicKey()),
			Addresses: []string{"foo"},
		},
	}

	p1b := &hyperspace.Announcement{
		Version: 1,
		Metadata: object.Metadata{
			Owner: peer.IDFromPublicKey(opk.PublicKey()),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Owner:     peer.IDFromPublicKey(opk.PublicKey()),
			Addresses: []string{"bar"},
		},
	}

	p2 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: peer.IDFromPublicKey(opk2.PublicKey()),
		},
		ConnectionInfo: &peer.ConnectionInfo{
			Owner:     peer.IDFromPublicKey(opk2.PublicKey()),
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
				Owner: peer.IDFromPublicKey(opk.PublicKey()),
			},
			ConnectionInfo: &peer.ConnectionInfo{
				Owner: peer.IDFromPublicKey(opk.PublicKey()),
			},
		},
		200*time.Millisecond,
	)

	pc.Remove(peer.IDFromPublicKey(opk.PublicKey()))

	pr, err := pc.Get(peer.IDFromPublicKey(opk.PublicKey()))
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
				Owner: peer.IDFromPublicKey(opk.PublicKey()),
			},
			ConnectionInfo: &peer.ConnectionInfo{
				Owner: peer.IDFromPublicKey(opk.PublicKey()),
			},
		},
		200*time.Millisecond,
	)

	time.Sleep(100 * time.Millisecond)

	pc.Touch(peer.IDFromPublicKey(opk.PublicKey()), 300*time.Millisecond)

	time.Sleep(200 * time.Millisecond)

	pr, err := pc.Get(peer.IDFromPublicKey(opk.PublicKey()))
	assert.NoError(t, err)
	assert.True(t, peer.IDFromPublicKey(opk.PublicKey()).Equals(
		pr.ConnectionInfo.Owner,
	))

	time.Sleep(300 * time.Millisecond)

	pr, err = pc.Get(peer.IDFromPublicKey(opk.PublicKey()))
	assert.Error(t, err)
	assert.Nil(t, pr)

	pc.Touch(peer.IDFromPublicKey(opk.PublicKey()), 300*time.Millisecond)
}

func TestPeerCache_TTL(t *testing.T) {
	opk, err := crypto.NewEd25519PrivateKey()
	assert.NoError(t, err)

	pc := NewPeerCache(200*time.Millisecond, "test4")

	pc.Put(
		&hyperspace.Announcement{
			Metadata: object.Metadata{
				Owner: peer.IDFromPublicKey(opk.PublicKey()),
			},
			ConnectionInfo: &peer.ConnectionInfo{
				Owner: peer.IDFromPublicKey(opk.PublicKey()),
			},
		},
		600*time.Millisecond,
	)

	pr, err := pc.Get(peer.IDFromPublicKey(opk.PublicKey()))
	assert.NoError(t, err)
	assert.True(t, peer.IDFromPublicKey(opk.PublicKey()).Equals(
		pr.ConnectionInfo.Owner,
	))

	time.Sleep(900 * time.Millisecond)

	pr, err = pc.Get(peer.IDFromPublicKey(opk.PublicKey()))
	assert.Error(t, err)
	assert.Nil(t, pr)
}
