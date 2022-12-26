package nimona

import (
	"testing"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/stretchr/testify/require"
)

func TestPublicKeyToBase58(t *testing.T) {
	pk0, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	b58 := PublicKeyToBase58(pk0)

	pk1, err := PublicKeyFromBase58(b58)
	require.NoError(t, err)

	require.True(t, pk0.Equal(pk1))
}
