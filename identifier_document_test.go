package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocumentID_Marshal(t *testing.T) {
	doc := &CborFixture{
		String: "foo",
	}

	hash, err := NewDocumentHash(doc)
	require.NoError(t, err)

	id := &DocumentID{
		DocumentHash: hash,
	}

	PrettyPrint(id)
}
