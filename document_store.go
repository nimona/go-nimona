package nimona

import (
	"errors"
	"fmt"
	"time"

	"github.com/geoah/go-pubsub"
	"gorm.io/gorm"

	"nimona.io/internal/kv"
	"nimona.io/tilde"
)

type (
	keyEdge struct {
		RootDocumentHash   DocumentHash
		ParentDocumentHash DocumentHash
		ChildDocumentHash  DocumentHash
	}
	vallueEdge struct {
		RootDocumentHash   DocumentHash
		ParentDocumentHash DocumentHash
		ChildDocumentHash  DocumentHash
		Sequence           uint64
	}
	keygraphIndex struct {
		RootDocumentHash DocumentHash
		DocumentHash     DocumentHash
	}
	keyTypeIndex struct {
		Type         string
		DocumentHash DocumentHash
	}
	keyAggregate struct {
		RootDocumentHash DocumentHash
		Path             string
		Key              string
	}
)

type DocumentStore struct {
	documents     kv.Store[DocumentHash, *Document]
	edges         kv.Store[keyEdge, *vallueEdge]
	types         kv.Store[keyTypeIndex, *DocumentHash]
	graphs        kv.Store[keygraphIndex, *DocumentHash]
	aggregates    kv.Store[keyAggregate, *[]byte]
	subscriptions *pubsub.Topic[*Document]
}

type (
	// DocumentEntry is a document entry in the database
	DocumentEntry struct {
		DocumentID       DocumentID  `gorm:"primaryKey"`
		DocumentType     string      `gorm:"index"`
		DocumentJSON     []byte      `gorm:"type:bytea"`
		RootDocumentHash *DocumentID `gorm:"index"`
		Sequence         uint64
		CreatedAt        time.Time `gorm:"autoCreateTime"`
	}
	DocumentEdge struct {
		RootDocumentHash   DocumentID `gorm:"primaryKey,index"`
		ParentDocumentHash DocumentID `gorm:"primaryKey,index"`
		ChildDocumentHash  DocumentID `gorm:"primaryKey,index"`
		Sequence           uint64
	}
)

func NewDocumentStore(db *gorm.DB) (*DocumentStore, error) {
	db = db.Debug()
	docStore, err := kv.NewSQLStore[DocumentHash, *Document](db, "documents")
	if err != nil {
		return nil, fmt.Errorf("error creating document store: %w", err)
	}

	edgeStore, err := kv.NewSQLStore[keyEdge, *vallueEdge](db, "edges")
	if err != nil {
		return nil, fmt.Errorf("error creating edge store: %w", err)
	}

	typeStore, err := kv.NewSQLStore[keyTypeIndex, *DocumentHash](db, "types")
	if err != nil {
		return nil, fmt.Errorf("error creating type store: %w", err)
	}

	graphStore, err := kv.NewSQLStore[keygraphIndex, *DocumentHash](db, "graphs")
	if err != nil {
		return nil, fmt.Errorf("error creating graph store: %w", err)
	}

	aggregateStore, err := kv.NewSQLStore[keyAggregate, *[]byte](db, "aggregates")
	if err != nil {
		return nil, fmt.Errorf("error creating graph store: %w", err)
	}

	s := &DocumentStore{
		documents:     docStore,
		edges:         edgeStore,
		types:         typeStore,
		graphs:        graphStore,
		aggregates:    aggregateStore,
		subscriptions: pubsub.NewTopic[*Document](),
	}

	return s, nil
}

func (s *DocumentStore) PutDocument(doc *Document) error {
	doc = doc.Copy()

	docID := NewDocumentID(doc)
	rootID := doc.Metadata.Root

	err := s.documents.Set(docID.DocumentHash, doc)
	if err != nil {
		return fmt.Errorf("error putting document: %w", err)
	}

	err = s.types.Set(
		keyTypeIndex{
			Type:         doc.Type(),
			DocumentHash: docID.DocumentHash,
		},
		&docID.DocumentHash,
	)
	if err != nil {
		return fmt.Errorf("error putting document type: %w", err)
	}

	if doc.Metadata.Root == nil {
		err = s.apply(doc)
		if err != nil {
			return fmt.Errorf("error applying document: %w", err)
		}

		go s.subscriptions.Publish(doc)
		return nil
	}

	err = s.graphs.Set(
		keygraphIndex{
			RootDocumentHash: rootID.DocumentHash,
			DocumentHash:     docID.DocumentHash,
		},
		&docID.DocumentHash,
	)
	if err != nil {
		return fmt.Errorf("error putting document graph: %w", err)
	}

	if len(doc.Metadata.Parents) == 0 {
		return errors.New("non root documents must have at least one parent")
	}

	for _, parentID := range doc.Metadata.Parents {
		if doc.Metadata.Sequence == 0 {
			return errors.New("non root documents must have a sequence number greater than 0")
		}
		err = s.edges.Set(
			keyEdge{
				RootDocumentHash:   rootID.DocumentHash,
				ParentDocumentHash: parentID.DocumentHash,
				ChildDocumentHash:  docID.DocumentHash,
			},
			&vallueEdge{
				RootDocumentHash:   rootID.DocumentHash,
				ParentDocumentHash: parentID.DocumentHash,
				ChildDocumentHash:  docID.DocumentHash,
				Sequence:           doc.Metadata.Sequence,
			},
		)
		if err != nil {
			return fmt.Errorf("error putting document: %w", err)
		}
	}

	err = s.apply(doc)
	if err != nil {
		return fmt.Errorf("error applying document: %w", err)
	}

	go s.subscriptions.Publish(doc)
	return nil
}

func (s *DocumentStore) GetDocument(docID DocumentID) (*Document, error) {
	m, err := s.documents.Get(docID.DocumentHash)
	if err != nil {
		return nil, fmt.Errorf("error getting document: %w", err)
	}

	return m, nil
}

// GetDocumentLeaves returns the leaves of a document graph, as well as the max sequence
// of the leaves.
func (s *DocumentStore) GetDocumentLeaves(docID DocumentID) ([]DocumentID, uint64, error) {
	edges, err := s.edges.GetPrefix(
		keyEdge{
			RootDocumentHash: docID.DocumentHash,
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting document: %w", err)
	}

	if len(edges) == 0 {
		// TODO should we check if we have the document first?
		return []DocumentID{docID}, 0, nil
	}

	// keep a list of all the parents
	parents := map[DocumentHash]struct{}{}
	for _, edge := range edges {
		parents[edge.ParentDocumentHash] = struct{}{}
	}

	var filtered []DocumentID
	var maxSeq uint64
	for _, edge := range edges {
		if edge.Sequence > maxSeq {
			maxSeq = edge.Sequence
		}
		if _, ok := parents[edge.ChildDocumentHash]; !ok {
			filtered = append(filtered, DocumentID{
				DocumentHash: edge.ChildDocumentHash,
			})
		}
	}

	return filtered, maxSeq, nil
}

func (s *DocumentStore) GetDocumentsByType(docType string) ([]*Document, error) {
	docHashes, err := s.types.GetPrefix(
		keyTypeIndex{
			Type: docType,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting documents from type index: %w", err)
	}

	var ret []*Document
	for _, docHash := range docHashes {
		m, err := s.documents.Get(*docHash)
		if err != nil {
			return nil, fmt.Errorf("error getting document: %w", err)
		}
		ret = append(ret, m)
	}

	return ret, nil
}

// GetDocumentsByRootID returns all documents with the given root id, not including
// the root document itself.
func (s *DocumentStore) GetDocumentsByRootID(id DocumentID) ([]*Document, error) {
	key := keygraphIndex{
		RootDocumentHash: id.DocumentHash,
	}
	docHashes, err := s.graphs.GetPrefix(key)
	if err != nil {
		return nil, fmt.Errorf("error getting documents: %w", err)
	}

	var docs []*Document
	for _, docHash := range docHashes {
		m, err := s.documents.Get(*docHash)
		if err != nil {
			return nil, fmt.Errorf("error getting document: %w", err)
		}
		docs = append(docs, m)
	}

	return docs, nil
}

// apply should only be called by PutDocument, after the document
// has been copied and stored.
func (s *DocumentStore) apply(doc *Document) error {
	// TODO: check if the document is already applied

	// if this is a root document, just insert it
	id := NewDocumentID(doc)
	if doc.Type() != "core/stream/patch" {
		j, err := doc.MarshalJSON()
		if err != nil {
			return fmt.Errorf("error marshaling document: %w", err)
		}
		err = s.aggregates.Set(
			keyAggregate{
				RootDocumentHash: id.DocumentHash,
			},
			&j,
		)
		if err != nil {
			return fmt.Errorf("error inserting aggregate: %w", err)
		}
		return nil
	}

	// TODO: validate the patch
	switch {
	case doc.Metadata.Root == nil:
		return errors.New("patch must have a root")
	}

	// retrieve root document and patches
	rootID := *doc.Metadata.Root
	rootDoc, err := s.GetDocument(rootID)
	if err != nil {
		return fmt.Errorf("error getting root document: %w", err)
	}
	patchDocs, err := s.GetDocumentsByRootID(rootID)
	if err != nil {
		return fmt.Errorf("error getting patches: %w", err)
	}

	// convert to patch objects
	var patches []*DocumentPatch
	for _, patchDoc := range patchDocs {
		patch := &DocumentPatch{}
		err = patch.FromDocument(patchDoc)
		if err != nil {
			return fmt.Errorf("error converting patch: %w", err)
		}
		patches = append(patches, patch)
	}

	// apply patches
	aggregateDoc, err := ApplyDocumentPatch(rootDoc, patches...)
	if err != nil {
		return fmt.Errorf("error applying patches: %w", err)
	}

	// replace aggregate
	j, err := aggregateDoc.MarshalJSON()
	if err != nil {
		return fmt.Errorf("error marshaling document: %w", err)
	}
	err = s.aggregates.Set(
		keyAggregate{
			RootDocumentHash: rootID.DocumentHash,
		},
		&j,
	)
	if err != nil {
		return fmt.Errorf("error inserting aggregate: %w", err)
	}

	return nil
}

func (s *DocumentStore) GetAggregate(
	rootID DocumentID,
) (*Document, error) {
	docBytes, err := s.aggregates.Get(
		keyAggregate{
			RootDocumentHash: rootID.DocumentHash,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting aggregate: %w", err)
	}

	doc := &Document{}
	err = doc.UnmarshalJSON(*docBytes)
	if err != nil {
		return nil, fmt.Errorf("error getting aggregate: %w", err)
	}

	return doc, nil
}

func (s *DocumentStore) CreatePatch(
	docID DocumentID,
	op string,
	path string,
	value tilde.Value,
	sctx SigningContext,
) (*Document, error) {
	switch op {
	case "replace", "append":
	default:
		return nil, fmt.Errorf("unsupported operation: %s", op)
	}

	// check if we have the document
	_, err := s.documents.Get(docID.DocumentHash)
	if err != nil {
		return nil, fmt.Errorf("error getting document: %w", err)
	}

	// get leaves and max sequence
	leaves, maxSeq, err := s.GetDocumentLeaves(docID)
	if err != nil {
		return nil, fmt.Errorf("error getting document leaves: %w", err)
	}

	patch := &DocumentPatch{
		Metadata: Metadata{
			Owner:     sctx.KeygraphID,
			Root:      &docID,
			Parents:   leaves,
			Sequence:  maxSeq + 1,
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		},
		Operations: []DocumentPatchOperation{{
			Op:    op,
			Path:  path,
			Value: value,
		}},
	}

	patchDoc := patch.Document()

	if !sctx.PrivateKey.IsZero() {
		sig := NewDocumentSignature(sctx.PrivateKey, NewDocumentHash(patchDoc))
		patchDoc.Metadata.Signature = sig
	}

	// TODO(geoah) should we store the doc here?

	return patchDoc, nil
}

func (s *DocumentStore) Subscribe(
	filters ...func(*Document) bool,
) *pubsub.Subscription[*Document] {
	return s.subscriptions.Subscribe(filters...)
}
