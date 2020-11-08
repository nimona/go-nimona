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

	p1 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: opk.PublicKey(),
		},
		QueryVector: hyperspace.New("foo", "bar"),
	}

	p2 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: opk2.PublicKey(),
		},
		QueryVector: hyperspace.New("foo", "not-bar"),
	}

	pc := NewPeerCache(200 * time.Millisecond)

	pc.Put(p1, 200*time.Millisecond)
	pc.Put(p2, 200*time.Millisecond)

	ps := pc.Lookup(hyperspace.New("foo"))
	assert.ElementsMatch(t, []*peer.Peer{p1, p2}, ps)

	ps = pc.Lookup(hyperspace.New("foo", "bar"))
	assert.ElementsMatch(t, []*peer.Peer{p1}, ps)

	ps = pc.Lookup(hyperspace.New("foo", "not-bar"))
	assert.ElementsMatch(t, []*peer.Peer{p2}, ps)
}

func TestPeerCache_List(t *testing.T) {
	opk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	opk2, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	p1a := &peer.Peer{
		Metadata: object.Metadata{
			Owner: opk.PublicKey(),
		},
		Addresses: []string{"foo"},
	}

	p1b := &peer.Peer{
		Metadata: object.Metadata{
			Owner: opk.PublicKey(),
		},
		Addresses: []string{"bar"},
	}

	p2 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: opk2.PublicKey(),
		},
		Addresses: []string{"foo"},
	}

	pc := NewPeerCache(200 * time.Millisecond)

	pc.Put(p1a, 200*time.Millisecond)
	pc.Put(p1b, 200*time.Millisecond)
	pc.Put(p2, 200*time.Millisecond)

	ps := pc.List()
	assert.ElementsMatch(t, []*peer.Peer{p1b, p2}, ps)
}

func TestPeerCache_Remove(t *testing.T) {
	opk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	pc := NewPeerCache(200 * time.Millisecond)

	pc.Put(
		&peer.Peer{
			Metadata: object.Metadata{
				Owner: opk.PublicKey(),
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

	pc := NewPeerCache(200 * time.Millisecond)

	pc.Put(
		&peer.Peer{
			Metadata: object.Metadata{
				Owner: opk.PublicKey(),
			},
		},
		200*time.Millisecond,
	)

	time.Sleep(100 * time.Millisecond)

	pc.Touch(opk.PublicKey(), 300*time.Millisecond)

	time.Sleep(200 * time.Millisecond)

	pr, err := pc.Get(opk.PublicKey())
	assert.NoError(t, err)
	assert.Equal(t, opk.PublicKey(), pr.Metadata.Owner)

	time.Sleep(300 * time.Millisecond)

	pr, err = pc.Get(opk.PublicKey())
	assert.Error(t, err)
	assert.Nil(t, pr)

	pc.Touch(opk.PublicKey(), 300*time.Millisecond)
}

func TestPeerCache_TTL(t *testing.T) {
	opk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	pc := NewPeerCache(200 * time.Millisecond)

	pc.Put(&peer.Peer{
		Metadata: object.Metadata{
			Owner: opk.PublicKey(),
		},
	}, 600*time.Millisecond)

	pr, err := pc.Get(opk.PublicKey())
	assert.NoError(t, err)
	assert.Equal(t, opk.PublicKey(), pr.Metadata.Owner)

	time.Sleep(700 * time.Millisecond)

	pr, err = pc.Get(opk.PublicKey())
	assert.Error(t, err)
	assert.Nil(t, pr)
}
