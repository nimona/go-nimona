package nimona

import (
	"testing"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/stretchr/testify/require"
)

func TestPeerID(t *testing.T) {
	pk, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	s0 := "nimona://peer:key:" + PublicKeyToBase58(pk)
	n0 := PeerID{
		PublicKey: pk,
	}

	require.Equal(t, s0, n0.String())

	n1, err := ParsePeerID(s0)
	require.NoError(t, err)
	require.Equal(t, n0, n1)
}
