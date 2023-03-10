package nimona

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDocumentHash_Ping(t *testing.T) {
	m := &documentFixture{
		String: "foo",
		Uint64: 42,
		Int64:  -42,
		Bytes:  []byte("bar"),
		Bool:   true,
		MapPtr: &documentFixture{
			String: "foo",
		},
		RepeatedString: []string{"foo", "bar"},
		RepeatedUint64: []uint64{1, 2, 3},
		RepeatedInt64:  []int64{-1, -2, -3},
		RepeatedBytes:  [][]byte{[]byte("foo"), []byte("bar")},
		// RepeatedBool:   []bool{true, false},
		RepeatedMap: []documentFixture{{
			String: "foo",
		}, {
			String: "bar",
		}},
	}

	exp := "8fp2d3WesUJ7xDzdN6SxFDZaWHvbXM8ejz9ETQM3mtrX"

	t.Run("hash", func(t *testing.T) {
		h := NewDocumentHash(m.Document())
		require.Equal(t, exp, h.String())
	})

	t.Run("unmarshal and hash", func(t *testing.T) {
		b, err := m.Document().MarshalJSON()
		require.NoError(t, err)

		fmt.Println(string(b))

		gm := &Document{}
		err = gm.UnmarshalJSON(b)
		require.NoError(t, err)

		g := &documentFixture{}
		err = g.FromDocument(gm)

		require.NoError(t, err)
		require.Equal(t, m, g)

		h := NewDocumentHash(g.Document())
		require.Equal(t, exp, h.String())
	})

	t.Run("ephemeral fields should not affect hash", func(t *testing.T) {
		m.EphemeralString = "foo"
		h := NewDocumentHash(m.Document())
		require.Equal(t, exp, h.String())
	})

	t.Run("add document id", func(t *testing.T) {
		m.DocumentID = NewDocumentID(m.Document())
		exp := "9fe24XevoiTEXrCcXbX2C72joHwpPms2x7kFpW2LiYJH"

		t.Run("test hashing", func(t *testing.T) {
			h := NewDocumentHash(m.Document())
			require.Equal(t, exp, h.String())
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

func TestDocumentHash(t *testing.T) {
	doc := &documentFixture{
		String: "foo",
	}

	hash := NewDocumentHash(doc.Document())
	require.Equal(t, "FBgb1MJcVCpPp2tLKJQp1R9E15a5Nntm9o93bB63XWSj", hash.String())
}

func TestDocumentHash_NewTestRandomDocumentID(t *testing.T) {
	t.Run("test random document id", func(t *testing.T) {
		h1 := NewTestRandomDocumentID(t)
		h2 := NewTestRandomDocumentID(t)
		require.NotEqual(t, h1, h2)
	})
}

// NewRandomHash returns a random hash for testing purposes.
func NewRandomHash(t *testing.T) DocumentHash {
	t.Helper()

	doc := &documentFixture{
		String: uuid.New().String(),
	}

	return NewDocumentHash(doc.Document())
}
