package nimona

import (
	"testing"

	"github.com/google/uuid"
)

func NewTestDocument(t *testing.T) *Document {
	t.Helper()
	doc := &CborFixture{
		String: uuid.New().String(),
	}
	return doc.Document()
}
