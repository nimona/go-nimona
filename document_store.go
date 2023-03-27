package nimona

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DocumentStore struct {
	db *gorm.DB
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
	s := &DocumentStore{
		db: db,
	}

	err := db.AutoMigrate(
		&DocumentEntry{},
		&DocumentEdge{},
	)
	if err != nil {
		return nil, fmt.Errorf("error migrating database: %w", err)
	}

	return s, nil
}

func (s *DocumentStore) PutDocument(doc *Document) error {
	docBytes, err := doc.MarshalJSON()
	if err != nil {
		return fmt.Errorf("error marshaling document: %w", err)
	}

	docID := NewDocumentID(doc)
	rootID := doc.Metadata.Root

	entry := &DocumentEntry{
		DocumentID:     docID,
		DocumentType:   doc.Type(),
		DocumentJSON:   docBytes,
		RootDocumentID: rootID,
		Sequence:       doc.Metadata.Sequence,
	}

	err = s.db.
		Clauses(
			clause.OnConflict{
				DoNothing: true,
			},
		).
		Create(entry).
		Error
	if err != nil {
		return fmt.Errorf("error putting document: %w", err)
	}

	// // for root documents, create a single edge to itself
	// // TODO should only consier patches as child documents?
	// // if doc.Type() != "core/stream/patch" {
	// if doc.Metadata.Root == nil && len(doc.Metadata.Parents) == 0 {
	// 	edge := &DocumentEdge{
	// 		RootDocumentID:   docID,
	// 		ParentDocumentID: docID,
	// 		ChildDocumentID:  docID,
	// 		Sequence:         0,
	// 	}
	// 	err = s.db.
	// 		Clauses(
	// 			clause.OnConflict{
	// 				DoNothing: true,
	// 			},
	// 		).
	// 		Create(edge).
	// 		Error
	// 	if err != nil {
	// 		return fmt.Errorf("error putting document: %w", err)
	// 	}
	// }

	if doc.Metadata.Root == nil {
		return nil
	}

	if len(doc.Metadata.Parents) == 0 {
		return errors.New("non root documents must have at least one parent")
	}

	edges := []*DocumentEdge{}
	for _, parentID := range doc.Metadata.Parents {
		if doc.Metadata.Sequence == 0 {
			return errors.New("non root documents must have a sequence number greater than 0")
		}
		edges = append(edges, &DocumentEdge{
			RootDocumentID:   *doc.Metadata.Root,
			ParentDocumentID: parentID,
			ChildDocumentID:  docID,
			Sequence:         doc.Metadata.Sequence,
		})
	}

	err = s.db.
		Clauses(
			clause.OnConflict{
				DoNothing: true,
			},
		).
		Create(edges).
		Error
	if err != nil {
		return fmt.Errorf("error putting document: %w", err)
	}

	return nil
}

func (s *DocumentStore) GetDocument(id DocumentID) (*Document, error) {
	doc := &DocumentEntry{}
	err := s.db.
		Where("document_id = ?", id).
		First(doc).
		Error
	if err != nil {
		return nil, fmt.Errorf("error getting document: %w", err)
	}

	m := &Document{}
	err = m.UnmarshalJSON(doc.DocumentJSON)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling document: %w", err)
	}

	return m, nil
}

// GetDocumentLeaves returns the leaves of a document graph, as well as the max sequence
// of the leaves.
func (s *DocumentStore) GetDocumentLeaves(id DocumentID) ([]DocumentID, uint64, error) {
	var edges []*DocumentEdge
	err := s.db.
		Where(`
			root_document_id = ?
			AND child_document_id NOT IN (
				SELECT DISTINCT parent_document_id
				FROM document_edges
				WHERE root_document_id = ?
			)`, id, id).
		Find(&edges).
		Error
	if err != nil {
		return nil, 0, fmt.Errorf("error getting document: %w", err)
	}

	if len(edges) == 0 {
		// TODO should we check if we have the document first?
		return []DocumentID{id}, 0, nil
	}

	var maxSeq uint64
	var ret []DocumentID
	for _, edge := range edges {
		ret = append(ret, edge.ChildDocumentID)
		if edge.Sequence > maxSeq {
			maxSeq = edge.Sequence
		}
	}

	return ret, maxSeq, nil
}

func (s *DocumentStore) GetDocumentsByType(docType string) ([]*Document, error) {
	var docs []*DocumentEntry
	err := s.db.
		Where("document_type = ?", docType).
		Find(&docs).
		Error
	if err != nil {
		return nil, fmt.Errorf("error getting documents: %w", err)
	}

	var ret []*Document
	for _, doc := range docs {
		m := &Document{}
		err = m.UnmarshalJSON(doc.DocumentJSON)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling document: %w", err)
		}
		ret = append(ret, m)
	}

	return ret, nil
}

// GetDocumentsByRootID returns all documents with the given root id, not including
// the root document itself.
func (s *DocumentStore) GetDocumentsByRootID(id DocumentID) ([]*Document, error) {
	var docs []*DocumentEntry
	err := s.db.
		Where("root_document_id = ?", id).
		Order("sequence ASC").
		Find(&docs).
		Error
	if err != nil {
		return nil, fmt.Errorf("error getting documents: %w", err)
	}

	var ret []*Document
	for _, doc := range docs {
		m := &Document{}
		err = m.UnmarshalJSON(doc.DocumentJSON)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling document: %w", err)
		}
		ret = append(ret, m)
	}

	return ret, nil
}

// nolint: unused // TODO: we should be using this probably, if not remove it
func gormErrUniqueViolation(err error) bool {
	e := errors.New("UNIQUE constraint failed")
	return !errors.Is(err, e)
}
