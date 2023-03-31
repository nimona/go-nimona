package nimona

import (
	"encoding/json"
	"errors"
	"fmt"
	reflect "reflect"
	"time"

	"gorm.io/gorm"

	"nimona.io/internal/kv"
)

type (
	keyEdge struct {
		RootDocumentID   DocumentID
		ParentDocumentID DocumentID
		ChildDocumentID  DocumentID
	}
	vallueEdge struct {
		RootDocumentID   DocumentID
		ParentDocumentID DocumentID
		ChildDocumentID  DocumentID
		Sequence         uint64
	}
	keyGraphIndex struct {
		RootDocumentID DocumentID
		DocumentID     DocumentID
	}
	keyTypeIndex struct {
		Type       string
		DocumentID DocumentID
	}
	keyAggregate struct {
		RootDocumentID DocumentID
		Path           string
		Key            string
	}
)

type DocumentStore struct {
	documents  kv.Store[DocumentID, Document]
	edges      kv.Store[keyEdge, vallueEdge]
	types      kv.Store[keyTypeIndex, DocumentID]
	graphs     kv.Store[keyGraphIndex, DocumentID]
	aggregates kv.Store[keyAggregate, []byte]
}

type (
	// DocumentEntry is a document entry in the database
	DocumentEntry struct {
		DocumentID     DocumentID  `gorm:"primaryKey"`
		DocumentType   string      `gorm:"index"`
		DocumentJSON   []byte      `gorm:"type:bytea"`
		RootDocumentID *DocumentID `gorm:"index"`
		Sequence       uint64
		CreatedAt      time.Time `gorm:"autoCreateTime"`
	}
	DocumentEdge struct {
		RootDocumentID   DocumentID `gorm:"primaryKey,index"`
		ParentDocumentID DocumentID `gorm:"primaryKey,index"`
		ChildDocumentID  DocumentID `gorm:"primaryKey,index"`
		Sequence         uint64
	}
)

func NewDocumentStore(db *gorm.DB) (*DocumentStore, error) {
	db = db.Debug()
	docStore, err := kv.NewSQLStore[DocumentID, Document](db, "documents")
	if err != nil {
		return nil, fmt.Errorf("error creating document store: %w", err)
	}

	edgeStore, err := kv.NewSQLStore[keyEdge, vallueEdge](db, "edges")
	if err != nil {
		return nil, fmt.Errorf("error creating edge store: %w", err)
	}

	typeStore, err := kv.NewSQLStore[keyTypeIndex, DocumentID](db, "types")
	if err != nil {
		return nil, fmt.Errorf("error creating type store: %w", err)
	}

	graphStore, err := kv.NewSQLStore[keyGraphIndex, DocumentID](db, "graphs")
	if err != nil {
		return nil, fmt.Errorf("error creating graph store: %w", err)
	}

	aggregateStore, err := kv.NewSQLStore[keyAggregate, []byte](db, "aggregates")
	if err != nil {
		return nil, fmt.Errorf("error creating graph store: %w", err)
	}

	s := &DocumentStore{
		documents:  docStore,
		edges:      edgeStore,
		types:      typeStore,
		graphs:     graphStore,
		aggregates: aggregateStore,
	}

	return s, nil
}

func (s *DocumentStore) PutDocument(doc *Document) error {
	docID := NewDocumentID(doc)
	rootID := doc.Metadata.Root

	err := s.documents.Set(docID, doc)
	if err != nil {
		return fmt.Errorf("error putting document: %w", err)
	}

	err = s.types.Set(
		keyTypeIndex{
			Type:       doc.Type(),
			DocumentID: docID,
		},
		&docID,
	)
	if err != nil {
		return fmt.Errorf("error putting document type: %w", err)
	}

	if doc.Metadata.Root == nil {
		return nil
	}

	err = s.graphs.Set(
		keyGraphIndex{
			RootDocumentID: *rootID,
			DocumentID:     docID,
		},
		&docID,
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
				RootDocumentID:   *rootID,
				ParentDocumentID: parentID,
				ChildDocumentID:  docID,
			},
			&vallueEdge{
				RootDocumentID:   *rootID,
				ParentDocumentID: parentID,
				ChildDocumentID:  docID,
				Sequence:         doc.Metadata.Sequence,
			},
		)
		if err != nil {
			return fmt.Errorf("error putting document: %w", err)
		}
	}

	return nil
}

func (s *DocumentStore) GetDocument(id DocumentID) (*Document, error) {
	m, err := s.documents.Get(id)
	if err != nil {
		return nil, fmt.Errorf("error getting document: %w", err)
	}

	return m, nil
}

// GetDocumentLeaves returns the leaves of a document graph, as well as the max sequence
// of the leaves.
func (s *DocumentStore) GetDocumentLeaves(id DocumentID) ([]DocumentID, uint64, error) {
	edges, err := s.edges.GetPrefix(
		keyEdge{
			RootDocumentID: id,
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting document: %w", err)
	}

	if len(edges) == 0 {
		// TODO should we check if we have the document first?
		return []DocumentID{id}, 0, nil
	}

	// keep a list of all the parents
	parents := map[DocumentID]struct{}{}
	for _, edge := range edges {
		parents[edge.ParentDocumentID] = struct{}{}
	}

	var filtered []DocumentID
	var maxSeq uint64
	for _, edge := range edges {
		if edge.Sequence > maxSeq {
			maxSeq = edge.Sequence
		}
		if _, ok := parents[edge.ChildDocumentID]; !ok {
			filtered = append(filtered, edge.ChildDocumentID)
		}
	}

	return filtered, maxSeq, nil
}

func (s *DocumentStore) GetDocumentsByType(docType string) ([]*Document, error) {
	docIDs, err := s.types.GetPrefix(
		keyTypeIndex{
			Type: docType,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting documents from type index: %w", err)
	}

	var ret []*Document
	for _, docID := range docIDs {
		m, err := s.documents.Get(*docID)
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
	key := keyGraphIndex{
		RootDocumentID: id,
	}
	docIDs, err := s.graphs.GetPrefix(key)
	if err != nil {
		return nil, fmt.Errorf("error getting documents: %w", err)
	}

	var docs []*Document
	for _, docID := range docIDs {
		m, err := s.documents.Get(*docID)
		if err != nil {
			return nil, fmt.Errorf("error getting document: %w", err)
		}
		docs = append(docs, m)
	}

	return docs, nil
}

func (s *DocumentStore) Apply(doc *Document) error {
	doc = doc.Copy()

	if doc.Type() != "core/stream/patch" {
		body, err := doc.MarshalJSON()
		if err != nil {
			return err
		}
		err = s.aggregates.Set(
			keyAggregate{
				RootDocumentID: NewDocumentID(doc),
			},
			&body,
		)
		if err != nil {
			return fmt.Errorf("error storing root doc: %w", err)
		}
		return nil
	}

	patch := &DocumentPatch{}
	err := patch.FromDocument(doc)
	if err != nil {
		return fmt.Errorf("error unmarshaling patch: %w", err)
	}

	for _, operation := range patch.Operations {
		switch operation.Op {
		case "replace":
			// TODO: implement
		case "append":
			body, err := json.Marshal(operation.Value)
			if err != nil {
				return fmt.Errorf("error marshaling op value: %w", err)
			}
			err = s.aggregates.Set(
				keyAggregate{
					RootDocumentID: *patch.Metadata.Root,
					Path:           operation.Path,
					Key:            operation.Key,
				},
				&body,
			)
			if err != nil {
				return fmt.Errorf("error storing op value: %w", err)
			}
		default:
			return fmt.Errorf("unsupported operation: %s", operation.Op)
		}
	}
	return nil
}

func (s *DocumentStore) GetAggregateNested(
	rootHash DocumentID,
	path string,
	target any, // *[]DocumentMapper,
) error {
	key := keyAggregate{
		RootDocumentID: rootHash,
		Path:           path,
	}
	bodies, err := s.aggregates.GetPrefix(key)
	if err != nil {
		return fmt.Errorf("error getting aggregate: %w", err)
	}

	// check if target is a pointer to a slice
	res := reflect.ValueOf(target).Elem()
	if res.Kind() != reflect.Slice {
		return fmt.Errorf("target is not a slice")
	}

	// check if target is a pointer to a slice of pointers
	if res.Type().Elem().Kind() != reflect.Ptr {
		return fmt.Errorf("target is not a slice of pointers")
	}

	// get the type of the slice elements
	typ := reflect.TypeOf(target).Elem().Elem().Elem()

	for _, body := range bodies {
		doc := &Document{}
		err := doc.UnmarshalJSON(*body)
		if err != nil {
			return fmt.Errorf("error unmarshaling aggregate: %w", err)
		}
		val := reflect.New(typ).Interface().(DocumentMapper)
		err = (val).FromDocument(doc)
		if err != nil {
			return fmt.Errorf("error unmarshaling aggregate: %w", err)
		}
		res.Set(reflect.Append(res, reflect.ValueOf(val)))
	}

	return nil
}
