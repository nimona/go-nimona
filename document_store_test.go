package nimona

import (
	"testing"

	_ "modernc.org/sqlite"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewTestDocumentDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(
		sqlite.Open("file::memory:?cache=shared"),
		&gorm.Config{},
	)
	require.NoError(t, err)

	return db
}

func NewTestDocumentStore(t *testing.T) *DocumentStore {
	t.Helper()

	store, err := NewDocumentStore(NewTestDocumentDB(t))
	require.NoError(t, err)

	return store
}

func TestDocumentStore(t *testing.T) {
	// Set up test DB
	store := NewTestDocumentStore(t)

	// Create a document
	doc := &CborFixture{
		String: "test",
		Uint64: 42,
	}

	docBytes, err := doc.MarshalCBORBytes()
	require.NoError(t, err)

	// Create an entry
	docID := NewDocumentID(doc)
	entry := &DocumentEntry{
		DocumentID:       docID,
		DocumentType:     "test",
		DocumentEncoding: "cbor",
		DocumentBytes:    docBytes,
		RootDocumentID: &DocumentID{
			DocumentHash: NewRandomHash(t),
		},
	}

	// Test putting a document
	err = store.PutDocument(entry)
	require.NoError(t, err)

	// Test getting an entry
	gotEntry, err := store.GetDocument(entry.DocumentID)
	require.NoError(t, err)
	require.True(t, gotEntry.DocumentID.IsEqual(docID))
	require.Equal(t, gotEntry.DocumentType, entry.DocumentType)
	require.Equal(t, gotEntry.DocumentEncoding, entry.DocumentEncoding)
	require.Equal(t, gotEntry.DocumentBytes, entry.DocumentBytes)
	require.NotNil(t, gotEntry.RootDocumentID)
	require.True(t, gotEntry.RootDocumentID.IsEqual(*entry.RootDocumentID))

	// Test unmarshaling the entry
	gotDoc := &CborFixture{}
	err = gotDoc.UnmarshalCBORBytes(gotEntry.DocumentBytes)
	require.NoError(t, err)
	require.Equal(t, gotDoc, doc)
}
