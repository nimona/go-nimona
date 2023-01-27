package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func NewTestKeyGraph(t *testing.T) *KeyGraph {
	t.Helper()

	pk, _, err := GenerateKey()
	require.NoError(t, err)

	return &KeyGraph{
		Keys: pk,
	}
}

func NewTestIdentity(t *testing.T) *Identity {
	t.Helper()

	return &Identity{
		KeyGraphID: NewDocumentID(NewTestKeyGraph(t)),
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
		b, err := MarshalCBORBytes(n0)
		require.NoError(t, err)

		n1 := &IdentityAlias{}
		err = UnmarshalCBORBytes(b, n1)
		require.NoError(t, err)
		require.EqualValues(t, n0, n1)
		require.Equal(t, s0, n1.String())
	})
}

func TestKeyGraph(t *testing.T) {
	kg := NewTestKeyGraph(t)

	t.Run("marshal unmarshal", func(t *testing.T) {
		b, err := MarshalCBORBytes(kg)
		require.NoError(t, err)

		kg1 := &KeyGraph{}
		err = UnmarshalCBORBytes(b, kg1)
		require.NoError(t, err)
		require.EqualValues(t, kg, kg1)
	})
}

func TestIdentity(t *testing.T) {
	id := NewTestIdentity(t)
	require.Equal(t, id.String(), id.String())

	t.Run("parse", func(t *testing.T) {
		id, err := ParseIdentity(id.String())
		require.NoError(t, err)
		require.Equal(t, id.String(), id.String())
	})

	t.Run("value/scan", func(t *testing.T) {
		val, err := id.Value()
		require.NoError(t, err)

		id := Identity{}
		err = id.Scan(val)
		require.NoError(t, err)
		require.Equal(t, id.String(), id.String())
	})
}
