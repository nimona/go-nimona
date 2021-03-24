package crypto

import (
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEd(t *testing.T) {
	p1, err := NewEd25519PrivateKey(PeerKey)

	t.Run("generate new private key", func(t *testing.T) {
		require.NoError(t, err)
		require.NotNil(t, p1)
	})

	t.Run("try marshaling/unmarshaling private key", func(t *testing.T) {
		p1s, err := p1.MarshalString()
		require.NoError(t, err)
		p1g := PrivateKey{}
		err = p1g.UnmarshalString(p1s)
		require.NoError(t, err)
		assert.Equal(t, p1, p1g)
	})

	t.Run("try marshaling/unmarshaling public key", func(t *testing.T) {
		r1 := p1.PublicKey()
		r1s, err := r1.MarshalString()
		require.NoError(t, err)
		r1g := PublicKey{}
		err = r1g.UnmarshalString(r1s)
		require.NoError(t, err)
		assert.Equal(t, r1, r1g)
		assert.True(t, r1.Equals(r1g))
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
