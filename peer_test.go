package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPeerKey(t *testing.T) {
	pk, _, err := GenerateKey()
	require.NoError(t, err)

	s0 := "nimona://peer:key:" + pk.String()
	n0 := &PeerKey{
		PublicKey: pk,
	}

	require.Equal(t, s0, n0.String())

	n1, err := ParsePeerKey(s0)
	require.NoError(t, err)
	require.Equal(t, n0, &n1)

	t.Run("marshal unmarshal", func(t *testing.T) {
		bc, err := MarshalCBORBytes(n0)
		require.NoError(t, err)

		n1 := &PeerKey{}
		err = UnmarshalCBORBytes(bc, n1)
		require.NoError(t, err)
		require.EqualValues(t, n0, n1)
		require.Equal(t, s0, n1.String())
	})
}
