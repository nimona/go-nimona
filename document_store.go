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
)

func NewDocumentStore(db *gorm.DB) (*DocumentStore, error) {
	s := &DocumentStore{
		db: db,
	}

	err := db.AutoMigrate(
		&DocumentEntry{},
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
