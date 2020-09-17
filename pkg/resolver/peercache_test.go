package resolver

import (
	"testing"
	"time"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"

	"github.com/stretchr/testify/assert"
)

func TestTTL(t *testing.T) {
	opk, err := crypto.GenerateEd25519PrivateKey()
	assert.NoError(t, err)

	pc := NewPeerCache(200 * time.Millisecond)

	pc.Put(&peer.Peer{
		Metadata: object.Metadata{
			Owner: opk.PublicKey(),
		},
	}, 600)

	pr, err := pc.Get(opk.PublicKey())
	assert.NoError(t, err)
	assert.Equal(t, opk.PublicKey(), pr.Metadata.Owner)

	time.Sleep(700 * time.Millisecond)

	pr, err = pc.Get(opk.PublicKey())
	assert.Error(t, err)
	assert.Nil(t, pr)
}
