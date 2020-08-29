package localpeer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/rand"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/peer"
)

func TestLocalPeer(t *testing.T) {
	k1, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	k2, err := crypto.GenerateEd25519PrivateKey()
	require.NoError(t, err)

	lp := New()

	lp.PutPrimaryPeerKey(k1)
	assert.Equal(t, k1, lp.GetPrimaryPeerKey())

	lp.PutPrimaryIdentityKey(k2)
	assert.Equal(t, k2, lp.GetPrimaryIdentityKey())

	ch1 := object.Hash("f01")
	ch2 := object.Hash("f02")

	lp.PutContentHashes(ch1)
	assert.ElementsMatch(t, []object.Hash{ch1}, lp.GetContentHashes())

	lp.PutContentHashes(ch1, ch2)
	assert.ElementsMatch(t, []object.Hash{ch1, ch2}, lp.GetContentHashes())

	lp.PutRelays(&peer.Peer{
		Metadata: object.Metadata{
			Owner: k1.PublicKey(),
		},
	})
	assert.ElementsMatch(t, []*peer.Peer{{
		Metadata: object.Metadata{
			Owner: k1.PublicKey(),
		},
	}}, lp.GetRelays())

	c1 := &peer.Certificate{
		Nonce: rand.String(6),
	}
	lp.PutCertificate(c1)
	assert.ElementsMatch(t, []*peer.Certificate{c1}, lp.GetCertificates())

	a1 := "foo"
	a2 := "foo2"
	lp.PutAddresses(a1, a2)
	assert.ElementsMatch(t, []string{a1, a2}, lp.GetAddresses())
}
