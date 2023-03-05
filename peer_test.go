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
		bc, err := n0.Document().MarshalJSON()
		require.NoError(t, err)

		m1 := &Document{}
		err = m1.UnmarshalJSON(bc)
		require.NoError(t, err)

		n1 := &PeerKey{}
		err = n1.FromDocumentMap(m1)
		require.NoError(t, err)
		require.EqualValues(t, n0, n1)
		require.Equal(t, s0, n1.String())
	})
}
