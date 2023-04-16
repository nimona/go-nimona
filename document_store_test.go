package nimona

import (
	"testing"
	"time"

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
	// Set up test store
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
	// Set up test store
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
	// Set up test store
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
	// Set up test store
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
	// Set up test store
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

func TestDocumentStore_Aggregate_Apply(t *testing.T) {
	store := NewTestDocumentStore(t)

	root := &Profile{
		DisplayName: "foo",
	}

	rootDoc := root.Document()
	rootDocID := NewDocumentID(rootDoc)

	patch1 := &DocumentPatch{
		Metadata: Metadata{
			Root: &rootDocID,
		},
		Operations: []DocumentPatchOperation{{
			Op:    "replace",
			Path:  "displayName",
			Value: tilde.String("bar"),
		}},
	}

	patch4doc := &ProfileRepository{
		Alias:  "testing.nimona.dev",
		Handle: "foo",
	}

	patch4 := &DocumentPatch{
		Metadata: Metadata{
			Root: &rootDocID,
		},
		Operations: []DocumentPatchOperation{{
			Op:        "append",
			Path:      "repositories",
			Value:     patch4doc.Map(),
			Key:       "9e81fff2-ec7a-44e2-aa05-617c84bbaf5e",
			Partition: []string{"2023", "04"},
		}},
	}

	store.Apply(rootDoc)
	store.Apply(patch1.Document())
	store.Apply(patch4.Document())

	res := []*ProfileRepository{}
	err := store.GetAggregateNested(
		rootDocID,
		"repositories",
		&res,
	)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))
	require.Equal(t, patch4doc, res[0])
}

func TestDocumentStore_CreatePatch(t *testing.T) {
	// Set up test store
	store := NewTestDocumentStore(t)

	idPtr := func(id DocumentID) *DocumentID {
		return &id
	}

	// Create a graph
	// One
	//  | \
	// Two Three
	//

	docOne := NewTestDocument(t)
	docOneID := idPtr(NewDocumentID(docOne))

	err := store.PutDocument(docOne)
	require.NoError(t, err)

	docTwo := NewTestDocument(t)
	docTwo.Metadata.Root = docOneID
	docTwo.Metadata.Parents = []DocumentID{*docOneID}
	docTwo.Metadata.Sequence = 1
	docTwoID := idPtr(NewDocumentID(docTwo))

	err = store.PutDocument(docTwo)
	require.NoError(t, err)

	docThree := NewTestDocument(t)
	docThree.Metadata.Root = docOneID
	docThree.Metadata.Parents = []DocumentID{*docOneID}
	docThree.Metadata.Sequence = 1
	docThreeID := idPtr(NewDocumentID(docThree))

	err = store.PutDocument(docThree)
	require.NoError(t, err)

	// Create a new keygraph, keys, and identity
	kg, kpc, _ := NewTestKeygraph(t)
	require.NoError(t, err)

	// Create a patch
	patchDoc, err := store.CreatePatch(
		*docOneID,
		"replace",
		"foo",
		tilde.String("bar"),
		SigningContext{
			KeygraphID: kg.ID(),
			PrivateKey: kpc.PrivateKey,
		},
	)
	require.NoError(t, err)

	expDoc := tilde.Map{
		"$type": tilde.String("core/stream/patch"),
		"$metadata": tilde.Map{
			"owner": kg.ID().TildeValue(),
			"root":  docOneID.Map(),
			"parents": tilde.List{
				docTwoID.Map(),
				docThreeID.Map(),
			},
			"sequence": tilde.Uint64(2),
		},
		"operations": tilde.List{
			tilde.Map{
				"op":    tilde.String("replace"),
				"path":  tilde.String("foo"),
				"value": tilde.String("bar"),
			},
		},
	}

	// Check timestamp and signature exist
	require.NotEmpty(t, patchDoc.Metadata.Timestamp)
	require.NotEmpty(t, patchDoc.Metadata.Signature)

	// And remove them for the comparison
	patchDoc.Metadata.Timestamp = ""
	patchDoc.Metadata.Signature = nil

	DumpDocument(patchDoc)

	require.EqualValues(t, expDoc, patchDoc.Map())
}

func TestDocumentStore_Subscribe(t *testing.T) {
	// Set up test store
	store := NewTestDocumentStore(t)

	docOne := NewTestDocument(t)

	sub := store.Subscribe(nil, func(doc *Document) bool {
		return doc.Type() == "test/fixture"
	})

	err := store.PutDocument(docOne)
	require.NoError(t, err)

	var gotDoc *Document

	select {
	case gotDoc = <-sub.Channel():
	case <-time.After(1 * time.Second):
		require.Fail(t, "timed out waiting for document")
	}

	EqualDocument(t, docOne, gotDoc)
}
