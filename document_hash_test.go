package nimona

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDocumentHash_Ping(t *testing.T) {
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

	exp := "82pCJU9HhpjB1XgndtpSzby9UT13dCmdbBVFnDyWysGq"

	t.Run("test marshaling", func(t *testing.T) {
		b, err := MarshalCBORBytes(m)
		require.NoError(t, err)

		t.Run("test hashing", func(t *testing.T) {
			h, err := NewDocumentHash(m)
			require.NoError(t, err)
			require.Equal(t, exp, h.String())
		})

		t.Run("test unmarshaling", func(t *testing.T) {
			g := &CborFixture{}
			err = UnmarshalCBORBytes(b, g)
			require.NoError(t, err)
			require.Equal(t, m, g)
		})
	})

	b, err := MarshalCBORBytes(m)
	require.NoError(t, err)

	t.Run("from cbor", func(t *testing.T) {
		h, err := NewDocumentHashFromCBOR(b)
		require.NoError(t, err)
		require.Equal(t, exp, h.String())
	})

	t.Run("from cborer", func(t *testing.T) {
		h, err := NewDocumentHash(m)
		require.NoError(t, err)
		require.Equal(t, exp, h.String())
	})

	t.Run("unmarshal and hash", func(t *testing.T) {
		g := &CborFixture{}
		err = UnmarshalCBORBytes(b, g)
		require.NoError(t, err)

		h, err := NewDocumentHash(g)
		require.NoError(t, err)
		require.Equal(t, exp, h.String())
	})

	t.Run("ephemeral fields should not affect hash", func(t *testing.T) {
		m.EphemeralString = "foo"
		h, err := NewDocumentHash(m)
		require.NoError(t, err)
		require.Equal(t, exp, h.String())
	})

	t.Run("add document id", func(t *testing.T) {
		m.DocumentID = NewDocumentIDFromCBOR(b)
		exp := "F31excad5AE5KndjRHFGx2gpU8XUrRuBiG19vLzC3LrK"

		t.Run("test hashing", func(t *testing.T) {
			h, err := NewDocumentHash(m)
			require.NoError(t, err)
			require.Equal(t, exp, h.String())
		})

		t.Run("test unmarshaling", func(t *testing.T) {
			g := &CborFixture{}
			err = UnmarshalCBORBytes(b, g)
			require.NoError(t, err)
		})
	})
}

func TestDocumentHash_NewRandomHash(t *testing.T) {
	t.Run("test random hash", func(t *testing.T) {
		h1 := NewRandomHash(t)
		h2 := NewRandomHash(t)
		require.NotEqual(t, h1, h2)
	})
}

// NewRandomHash returns a random hash for testing purposes.
func NewRandomHash(t *testing.T) DocumentHash {
	t.Helper()

	doc := &CborFixture{
		String: uuid.New().String(),
	}

	hash, err := NewDocumentHash(doc)
	require.NoError(t, err)

	return hash
}
