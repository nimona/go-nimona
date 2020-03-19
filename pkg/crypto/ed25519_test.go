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

	fmt.Println("private key", p1)
	fmt.Println("public key", p1.PublicKey())

	r1 := p1.PublicKey()
	rA, err := parse25519PublicKey(p1.PublicKey().String())
	require.NoError(t, err)
	require.Equal(t, r1, NewPublicKey(rA))

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
