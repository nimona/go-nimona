package peerstore

import (
	"testing"
	"time"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/hyperspace"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"

	"github.com/stretchr/testify/assert"
)

func TestPeerCache_Lookup(t *testing.T) {
	opk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	opk2, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	p1 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: opk.PublicKey(),
		},
		Peer: &peer.ConnectionInfo{
			PublicKey: opk.PublicKey(),
		},
		PeerVector: hyperspace.New("foo", "bar"),
	}

	p2 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: opk2.PublicKey(),
		},
		Peer: &peer.ConnectionInfo{
			PublicKey: opk2.PublicKey(),
		},
		PeerVector: hyperspace.New("foo", "not-bar"),
	}

	pc := NewPeerCache(200*time.Millisecond, "test0")

	pc.Put(p1, 200*time.Millisecond)
	pc.Put(p2, 200*time.Millisecond)

	ps := pc.Lookup(hyperspace.New("foo"))
	assert.ElementsMatch(t, []*hyperspace.Announcement{p1, p2}, ps)

	ps = pc.Lookup(hyperspace.New("foo", "bar"))
	assert.ElementsMatch(t, []*hyperspace.Announcement{p1}, ps)

	ps = pc.Lookup(hyperspace.New("foo", "not-bar"))
	assert.ElementsMatch(t, []*hyperspace.Announcement{p2}, ps)
}

func TestPeerCache_List(t *testing.T) {
	opk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	opk2, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	p1a := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: opk.PublicKey(),
		},
		Peer: &peer.ConnectionInfo{
			PublicKey: opk.PublicKey(),
			Addresses: []string{"foo"},
		},
	}

	p1b := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: opk.PublicKey(),
		},
		Peer: &peer.ConnectionInfo{
			PublicKey: opk.PublicKey(),
			Addresses: []string{"bar"},
		},
	}

	p2 := &hyperspace.Announcement{
		Metadata: object.Metadata{
			Owner: opk2.PublicKey(),
		},
		Peer: &peer.ConnectionInfo{
			PublicKey: opk2.PublicKey(),
			Addresses: []string{"foo"},
		},
	}

	pc := NewPeerCache(200*time.Millisecond, "test1")

	pc.Put(p1a, 200*time.Millisecond)
	pc.Put(p1b, 200*time.Millisecond)
	pc.Put(p2, 200*time.Millisecond)

	ps := pc.List()
	assert.ElementsMatch(t, []*hyperspace.Announcement{p1b, p2}, ps)
}

func TestPeerCache_Remove(t *testing.T) {
	opk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	pc := NewPeerCache(200*time.Millisecond, "test2")

	pc.Put(
		&hyperspace.Announcement{
			Metadata: object.Metadata{
				Owner: opk.PublicKey(),
			},
			Peer: &peer.ConnectionInfo{
				PublicKey: opk.PublicKey(),
			},
		},
		200*time.Millisecond,
	)

	pc.Remove(opk.PublicKey())

	pr, err := pc.Get(opk.PublicKey())
	assert.Error(t, err)
	assert.Nil(t, pr)
}

func TestPeerCache_Touch(t *testing.T) {
	opk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	pc := NewPeerCache(200*time.Millisecond, "test3")

	pc.Put(
		&hyperspace.Announcement{
			Metadata: object.Metadata{
				Owner: opk.PublicKey(),
			},
			Peer: &peer.ConnectionInfo{
				PublicKey: opk.PublicKey(),
			},
		},
		200*time.Millisecond,
	)

	time.Sleep(100 * time.Millisecond)

	pc.Touch(opk.PublicKey(), 300*time.Millisecond)

	time.Sleep(200 * time.Millisecond)

	pr, err := pc.Get(opk.PublicKey())
	assert.NoError(t, err)
	assert.Equal(t, opk.PublicKey(), pr.Peer.PublicKey)

	time.Sleep(300 * time.Millisecond)

	pr, err = pc.Get(opk.PublicKey())
	assert.Error(t, err)
	assert.Nil(t, pr)

	pc.Touch(opk.PublicKey(), 300*time.Millisecond)
}

func TestPeerCache_TTL(t *testing.T) {
	opk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	pc := NewPeerCache(200*time.Millisecond, "test4")

	pc.Put(
		&hyperspace.Announcement{
			Metadata: object.Metadata{
				Owner: opk.PublicKey(),
			},
			Peer: &peer.ConnectionInfo{
				PublicKey: opk.PublicKey(),
			},
		},
		600*time.Millisecond,
	)

	pr, err := pc.Get(opk.PublicKey())
	assert.NoError(t, err)
	assert.Equal(t, opk.PublicKey(), pr.Peer.PublicKey)

	time.Sleep(900 * time.Millisecond)

	pr, err = pc.Get(opk.PublicKey())
	assert.Error(t, err)
	assert.Nil(t, pr)
}
