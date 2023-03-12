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
	val := &documentFixture{
		String: "test",
		Uint64: 42,
	}
	doc := val.Document()

	docBytes, err := doc.MarshalJSON()
	require.NoError(t, err)

	// Create an entry
	docID := NewDocumentID(doc)
	entry := &DocumentEntry{
		DocumentID:   docID,
		DocumentType: "test/fixture",
		DocumentJSON: docBytes,
	}

	// Test putting a document
	err = store.PutDocument(doc)
	require.NoError(t, err)

	// Test getting an entry
	gotDoc, err := store.GetDocument(entry.DocumentID)
	require.NoError(t, err)
	require.EqualValues(t, gotDoc, doc)
}

func TestDocumentStore_GetDocumentsByRootID(t *testing.T) {
	// Set up test DB
	store := NewTestDocumentStore(t)

	// Create an entry
	rootDoc := NewTestDocument(t)
	childDoc := NewTestDocument(t)
	rootDocID := NewDocumentID(rootDoc)
	childDoc.Metadata.Root = &rootDocID

	err := store.PutDocument(rootDoc)
	require.NoError(t, err)

	err = store.PutDocument(childDoc)
	require.NoError(t, err)

	// Test getting the stream
	gotEntries, err := store.GetDocumentsByRootID(rootDocID)
	require.NoError(t, err)
	require.Len(t, gotEntries, 1)

	EqualDocument(t, childDoc, gotEntries[0])
}
