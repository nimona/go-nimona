package nimona

import (
	"testing"

	"github.com/google/uuid"
)

// NewTestRandomDocumentID returns a random document ID for testing purposes.
func NewTestRandomDocumentID(t *testing.T) DocumentID {
	t.Helper()

	doc := &CborFixture{
		String: uuid.New().String(),
	}

	return NewDocumentID(doc.Document())
}
