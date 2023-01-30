package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublicKeyToBase58(t *testing.T) {
	pk0, _, err := GenerateKey()
	require.NoError(t, err)

	b58 := pk0.String()

	pk1, err := ParsePublicKey(b58)
	require.NoError(t, err)

	require.True(t, pk0.Equal(pk1))
}
