package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityAlias(t *testing.T) {
	s0 := "nimona://id:alias:testing.romdo.io/geoah"
	n0 := &IdentityAlias{
		Hostname: "testing.romdo.io",
		Path:     "geoah",
	}

	require.Equal(t, s0, n0.String())

	n1, err := ParseIdentityAlias(s0)
	require.NoError(t, err)
	require.Equal(t, n0, n1)

	t.Run("marshal unmarshal", func(t *testing.T) {
		n1 := &IdentityAlias{}
		n1.FromDocument(n0.Document())
		require.NoError(t, err)
		require.EqualValues(t, n0, n1)
		require.Equal(t, s0, n1.String())
	})
}
