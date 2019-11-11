package crypto

import (
	"crypto/rand"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEd(t *testing.T) {
	p1, err := GenerateEd25519PrivateKey()
	require.NoError(t, err)
	require.NotNil(t, p1)

	pp1 := p1.PublicKey().ToObject()
	var ppA PublicKey
	err = ppA.FromObject(pp1)
	require.NoError(t, err)
	require.Equal(t, p1.PublicKey(), ppA)

	fmt.Println("private key", p1)
	fmt.Println("public key", p1.PublicKey())

	r1 := p1.PublicKey()
	rA, err := parse25519PublicKey(p1.PublicKey().String())
	require.NoError(t, err)
	require.Equal(t, r1, NewPublicKey(rA))

	// p2, err := GenerateEd25519PrivateKey()
	// require.NoError(t, err)
	// require.NotNil(t, p2)

	// p3, err := GenerateEd25519PrivateKey()
	// require.NoError(t, err)
	// require.NotNil(t, p3)

	// s1 := p1.Shared(&p2.PublicKey)
	// s2 := p2.Shared(&p1.PublicKey)
	// require.Equal(t, s1, s2)

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

	// sX := p3.Shared(&p1.PublicKey)
	// require.NotEqual(t, s1, sX)
	// require.NotEqual(t, s2, sX)
}
