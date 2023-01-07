package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageHash_Ping(t *testing.T) {
	m := &CborFixture{
		String: "foo",
		Uint64: 42,
		Int64:  -42,
		Bytes:  []byte("bar"),
		Bool:   true,
		Map: &CborFixture{
			String: "foo",
		},
		RepeatedString: []string{"foo", "bar"},
		RepeatedUint64: []uint64{1, 2, 3},
		RepeatedInt64:  []int64{-1, -2, -3},
		RepeatedBytes:  [][]byte{[]byte("foo"), []byte("bar")},
		// RepeatedBool:   []bool{true, false},
		RepeatedMap: []*CborFixture{{
			String: "foo",
		}, {
			String: "bar",
		}},
	}

	exp := "ANYEibdUhncK5kumasnV7Q3FsF4PBpL1GbyiZd3QN1XA"

	t.Run("test marshaling", func(t *testing.T) {
		b, err := m.MarshalCBORBytes()
		require.NoError(t, err)

		g := &CborFixture{}
		err = g.UnmarshalCBORBytes(b)
		require.NoError(t, err)

		require.Equal(t, m, g)
	})

	b, err := m.MarshalCBORBytes()
	require.NoError(t, err)

	t.Run("from cbor", func(t *testing.T) {
		h, err := MessageHashFromCBOR(b)
		require.NoError(t, err)
		require.Equal(t, exp, h.String())
	})

	t.Run("from cborer", func(t *testing.T) {
		h, err := MessageHash(m)
		require.NoError(t, err)
		require.Equal(t, exp, h.String())
	})

	t.Run("unmarshal and hash", func(t *testing.T) {
		g := &CborFixture{}
		err = g.UnmarshalCBORBytes(b)
		require.NoError(t, err)

		h, err := MessageHash(g)
		require.NoError(t, err)
		require.Equal(t, exp, h.String())
	})

	t.Run("ephemeral fields should not affect hash", func(t *testing.T) {
		m.EphemeralString = "foo"
		h, err := MessageHash(m)
		require.NoError(t, err)
		require.Equal(t, exp, h.String())
	})
}
