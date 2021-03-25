package crypto

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEd(t *testing.T) {
	p1, err := GenerateEd25519PrivateKey(PeerKey)

	require.Equal(t, PeerKey, p1.Type())
	require.Equal(t, PeerKey, p1.PublicKey().Type())

	t.Run("generate new private key", func(t *testing.T) {
		require.NoError(t, err)
		require.NotNil(t, p1)
	})

	t.Run("try encoding/decoding private key", func(t *testing.T) {
		rp1, err := ed25519PrivateFromPrivateKey(p1)
		assert.NoError(t, err)
		assert.Equal(t, p1.ed25519(), rp1)
	})

	t.Run("try encoding/decoding public key", func(t *testing.T) {
		r1 := p1.PublicKey()
		rr1, err := ed25519PublicFromPublicKey(r1)
		assert.NoError(t, err)
		assert.Equal(t, r1.ed25519(), rr1)
	})

	b := make([]byte, 5647)
	n, err := io.ReadFull(rand.Reader, b)
	require.NoError(t, err)
	require.Equal(t, 5647, n)

	s1 := p1.Sign(b)
	require.NotEmpty(t, s1)

	err = p1.PublicKey().Verify(b, s1)
	require.NoError(t, err)

	s1[0] = 0
	err = p1.PublicKey().Verify(b, s1)
	require.Error(t, err)
}
