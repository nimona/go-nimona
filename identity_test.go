package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func NewTestIdentity(t *testing.T) *Identity {
	t.Helper()

	pk, _, err := GenerateKey()
	require.NoError(t, err)

	return &Identity{
		Keys: pk,
	}
}

func TestIdentityAlias(t *testing.T) {
	s0 := "nimona://id:alias:testing.romdo.io/geoah"
	n0 := &IdentityAlias{
		Network: NetworkAlias{
			Hostname: "testing.romdo.io",
		},
		Handle: "geoah",
	}

	require.Equal(t, s0, n0.String())

	n1, err := ParseIdentityAlias(s0)
	require.NoError(t, err)
	require.Equal(t, n0, n1)

	t.Run("marshal unmarshal", func(t *testing.T) {
		b, err := n0.MarshalCBORBytes()
		require.NoError(t, err)

		n1 := &IdentityAlias{}
		err = n1.UnmarshalCBORBytes(b)
		require.NoError(t, err)
		require.EqualValues(t, n0, n1)
		require.Equal(t, s0, n1.String())
	})
}

func TestIdentity(t *testing.T) {
	id := NewTestIdentity(t)

	t.Run("marshal unmarshal", func(t *testing.T) {
		b, err := id.MarshalCBORBytes()
		require.NoError(t, err)

		id1 := &Identity{}
		err = id1.UnmarshalCBORBytes(b)
		require.NoError(t, err)
		require.EqualValues(t, id, id1)
		require.Equal(t, id.String(), id1.String())
	})

	t.Run("IdentityID", func(t *testing.T) {
		idID := id.IdentityID()
		require.Equal(t, id.String(), idID.String())
	})
}

func TestIdentityID(t *testing.T) {
	id := NewTestIdentity(t)
	idID := id.IdentityID()
	require.Equal(t, id.String(), idID.String())

	t.Run("parse", func(t *testing.T) {
		idID, err := ParseIdentityID(idID.String())
		require.NoError(t, err)
		require.Equal(t, id.String(), idID.String())
	})

	t.Run("value/scan", func(t *testing.T) {
		val, err := idID.Value()
		require.NoError(t, err)

		var idID IdentityID
		err = idID.Scan(val)
		require.NoError(t, err)
		require.Equal(t, id.String(), idID.String())
	})
}
