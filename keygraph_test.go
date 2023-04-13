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

func NewTestKeyGraphID(t *testing.T) KeyGraphID {
	t.Helper()

	kg, _, _ := NewTestKeyGraph(t)
	return kg.ID()
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

func TestKeyGraphID(t *testing.T) {
	kg, _, _ := NewTestKeyGraph(t)
	id := kg.ID()

	t.Run("test id", func(t *testing.T) {
		require.NotEmpty(t, id)
		require.NotEmpty(t, id.String())
	})

	t.Run("test id value/scan", func(t *testing.T) {
		val, err := id.Value()
		require.NoError(t, err)

		id := KeyGraphID{}
		err = id.Scan(val)
		require.NoError(t, err)
		require.Equal(t, id.String(), id.String())
	})
}

func TestKeyGraphID_ParseKeyGraphNRI(t *testing.T) {
	id := NewTestKeyGraphID(t)
	got, err := ParseKeyGraphNRI(id.NRI())
	require.NoError(t, err)
	require.Equal(t, id, got)
}
