package nimona

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPeerID(t *testing.T) {
	pk, _, err := GenerateKey()
	require.NoError(t, err)

	s0 := "nimona://peer:key:" + pk.String()
	n0 := PeerID{
		PublicKey: pk,
	}

	require.Equal(t, s0, n0.String())

	n1, err := ParsePeerID(s0)
	require.NoError(t, err)
	require.Equal(t, n0, n1)

	bc, err := n0.MarshalCBORBytes()
	require.NoError(t, err)
	fmt.Printf("%x\n", bc)
}
