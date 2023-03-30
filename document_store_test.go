package nimona

import (
	"testing"

	_ "modernc.org/sqlite"

	"nimona.io/tilde"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewTestDocumentDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(
		sqlite.Open("file::memory:"),
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
	childDoc.Metadata.Parents = []DocumentID{NewDocumentID(childDoc)}
	childDoc.Metadata.Sequence = 1

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

func TestDocumentStore_GetDocumentsByType(t *testing.T) {
	// Set up test DB
	store := NewTestDocumentStore(t)

	// Create documents
	doc1 := NewDocument(
		tilde.Map{
			"$type": tilde.String("type-a"),
			"foo":   tilde.String("foo"),
		},
	)
	doc2 := NewDocument(
		tilde.Map{
			"$type": tilde.String("type-a"),
			"foo":   tilde.String("bar"),
		},
	)
	doc3 := NewDocument(
		tilde.Map{
			"$type": tilde.String("type-b"),
			"foo":   tilde.String("baz"),
		},
	)

	err := store.PutDocument(doc1)
	require.NoError(t, err)

	err = store.PutDocument(doc2)
	require.NoError(t, err)

	err = store.PutDocument(doc3)
	require.NoError(t, err)

	// Test getting by type
	gotEntries, err := store.GetDocumentsByType("type-a")
	require.NoError(t, err)
	require.Len(t, gotEntries, 2)

	EqualDocument(t, doc1, gotEntries[0])
	EqualDocument(t, doc2, gotEntries[1])
}

func TestDocumentStore_GetDocumentLeaves_Empty(t *testing.T) {
	// Set up test DB
	store := NewTestDocumentStore(t)

	idPtr := func(id DocumentID) *DocumentID {
		return &id
	}

	// Create a graph
	// One

	docOne := NewTestDocument(t)
	docOneID := idPtr(NewDocumentID(docOne))

	err := store.PutDocument(docOne)
	require.NoError(t, err)

	// Test getting the stream
	gotEntries, seq, err := store.GetDocumentLeaves(*docOneID)
	require.NoError(t, err)
	require.Equal(t, uint64(0), seq)
	require.Len(t, gotEntries, 1)

	require.EqualValues(t, []DocumentID{
		*docOneID,
	}, gotEntries)
}

func TestDocumentStore_GetDocumentLeaves(t *testing.T) {
	// Set up test DB
	store := NewTestDocumentStore(t)

	idPtr := func(id DocumentID) *DocumentID {
		return &id
	}

	// Create a graph
	// One
	//  | \
	// Two Three
	//  |
	// Four
	//

	docOne := NewTestDocument(t)
	docOneID := idPtr(NewDocumentID(docOne))

	err := store.PutDocument(docOne)
	require.NoError(t, err)

	docTwo := NewTestDocument(t)
	docTwo.Metadata.Root = docOneID
	docTwo.Metadata.Parents = []DocumentID{*docOneID}
	docTwo.Metadata.Sequence = 1

	err = store.PutDocument(docTwo)
	require.NoError(t, err)

	docThree := NewTestDocument(t)
	docThree.Metadata.Root = docOneID
	docThree.Metadata.Parents = []DocumentID{*docOneID}
	docThree.Metadata.Sequence = 1

	err = store.PutDocument(docThree)
	require.NoError(t, err)

	docFour := NewTestDocument(t)
	docFour.Metadata.Root = docOneID
	docFour.Metadata.Parents = []DocumentID{NewDocumentID(docTwo)}
	docFour.Metadata.Sequence = 2

	err = store.PutDocument(docFour)
	require.NoError(t, err)

	// Test getting the stream
	gotEntries, seq, err := store.GetDocumentLeaves(*docOneID)
	require.NoError(t, err)
	require.Equal(t, uint64(2), seq)
	require.Len(t, gotEntries, 2)

	require.EqualValues(t, []DocumentID{
		NewDocumentID(docThree),
		NewDocumentID(docFour),
	}, gotEntries)
}
