package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func NewTestKeygraph(t *testing.T) (*Keygraph, *KeyPair, *KeyPair) {
	t.Helper()

	kp1, err := GenerateKeyPair()
	require.NoError(t, err)

	kp2, err := GenerateKeyPair()
	require.NoError(t, err)

	return NewKeygraph(kp1.PublicKey, kp2.PublicKey), kp1, kp2
}

func NewTestKeygraphID(t *testing.T) KeygraphID {
	t.Helper()

	kg, _, _ := NewTestKeygraph(t)
	return kg.ID()
}

func TestKeygraph(t *testing.T) {
	kg, _, _ := NewTestKeygraph(t)

	t.Run("test hash", func(t *testing.T) {
		h := NewDocumentHash(kg.Document())
		require.NotEmpty(t, h)
	})

	t.Run("marshal unmarshal", func(t *testing.T) {
		kg1 := &Keygraph{}
		kg1.FromDocument(kg.Document())
		require.EqualValues(t, kg, kg1)
	})
}

func TestKeygraphID(t *testing.T) {
	kg, _, _ := NewTestKeygraph(t)
	id := kg.ID()

	t.Run("test id", func(t *testing.T) {
		require.NotEmpty(t, id)
		require.NotEmpty(t, id.String())
	})

	t.Run("test id value/scan", func(t *testing.T) {
		val, err := id.Value()
		require.NoError(t, err)

		id := KeygraphID{}
		err = id.Scan(val)
		require.NoError(t, err)
		require.Equal(t, id.String(), id.String())
	})
}

func TestKeygraphID_ParseKeygraphNRI(t *testing.T) {
	id := NewTestKeygraphID(t)
	got, err := ParseKeygraphNRI(id.NRI())
	require.NoError(t, err)
	require.Equal(t, id, got)
}
