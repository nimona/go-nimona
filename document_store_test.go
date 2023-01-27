package nimona

import (
	"testing"
	"time"

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

	docBytes, err := MarshalCBORBytes(doc)
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
	err = UnmarshalCBORBytes(gotEntry.DocumentBytes, gotDoc)
	require.NoError(t, err)
	require.Equal(t, gotDoc, doc)
}

func TestDocumentStore_GetDocumentsByRootID(t *testing.T) {
	// Set up test DB
	store := NewTestDocumentStore(t)

	// Create an entry
	rootEntry := &DocumentEntry{
		DocumentID:       NewTestRandomDocumentID(t),
		DocumentType:     "test",
		DocumentEncoding: "cbor",
		DocumentBytes:    []byte("root"),
		Sequence:         0,
	}
	childEntry := &DocumentEntry{
		DocumentID:       NewTestRandomDocumentID(t),
		DocumentType:     "test",
		DocumentEncoding: "cbor",
		DocumentBytes:    []byte("child"),
		RootDocumentID:   &rootEntry.DocumentID,
		Sequence:         1,
	}

	err := store.PutDocument(rootEntry)
	require.NoError(t, err)

	err = store.PutDocument(childEntry)
	require.NoError(t, err)

	// Test getting the stream
	gotEntries, err := store.GetDocumentsByRootID(rootEntry.DocumentID)
	require.NoError(t, err)
	require.Len(t, gotEntries, 1)

	// Ignore the datetimes
	rootEntry.CreatedAt = time.Time{}
	childEntry.CreatedAt = time.Time{}
	for _, e := range gotEntries {
		e.CreatedAt = time.Time{}
	}

	require.EqualValues(t, childEntry, gotEntries[0])
}
