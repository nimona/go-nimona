package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func NewTestKeyGraph(t *testing.T) (*KeyGraph, *KeyPair, *KeyPair) {
	t.Helper()

	kp1, err := GenerateKeyPair()
	require.NoError(t, err)

	kp2, err := GenerateKeyPair()
	require.NoError(t, err)

	return NewKeyGraph(kp1.PublicKey, kp2.PublicKey), kp1, kp2
}

func NewTestIdentity(t *testing.T) *Identity {
	t.Helper()

	kg, _, _ := NewTestKeyGraph(t)
	return &Identity{
		KeyGraph: NewDocumentHash(kg.Document()),
	}
}

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

func TestKeyGraph(t *testing.T) {
	kg, _, _ := NewTestKeyGraph(t)

	t.Run("test hash", func(t *testing.T) {
		h := NewDocumentHash(kg.Document())
		require.NotEmpty(t, h)
	})

	t.Run("marshal unmarshal", func(t *testing.T) {
		kg1 := &KeyGraph{}
		kg1.FromDocument(kg.Document())
		require.EqualValues(t, kg, kg1)
	})
}

func TestIdentity(t *testing.T) {
	id := NewTestIdentity(t)
	require.Equal(t, id.String(), id.String())

	t.Run("value/scan", func(t *testing.T) {
		val, err := id.Value()
		require.NoError(t, err)

		id := Identity{}
		err = id.Scan(val)
		require.NoError(t, err)
		require.Equal(t, id.String(), id.String())
	})
}
