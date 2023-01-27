package nimona

import (
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
		DocumentID       DocumentID `gorm:"primaryKey"`
		DocumentType     string     `gorm:"index"`
		DocumentEncoding string
		DocumentBytes    []byte      `gorm:"type:bytea"`
		RootDocumentID   *DocumentID `gorm:"index"`
		Sequence         uint64
		CreatedAt        time.Time `gorm:"autoCreateTime"`
	}
)

func (doc *DocumentEntry) UnmarshalInto(v Cborer) error {
	err := UnmarshalCBORBytes(doc.DocumentBytes, v)
	if err != nil {
		return fmt.Errorf("error unmarshaling document: %w", err)
	}
	return nil
}

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

// PutDocument puts a document in the database, if it already exists, return
// a ErrDocumentAlreadyExists error.
func (s *DocumentStore) PutDocument(entry *DocumentEntry) error {
	if entry.DocumentID.IsEmpty() {
		return fmt.Errorf("document id is empty")
	}
	if entry.DocumentType == "" {
		return fmt.Errorf("document type is empty")
	}
	if entry.DocumentEncoding == "" {
		return fmt.Errorf("document encoding is empty")
	}
	if len(entry.DocumentBytes) == 0 {
		return fmt.Errorf("document bytes is empty")
	}
	err := s.db.
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

func (s *DocumentStore) GetDocument(id DocumentID) (*DocumentEntry, error) {
	doc := &DocumentEntry{}
	err := s.db.
		Where("document_id = ?", id).
		First(doc).
		Error
	if err != nil {
		return nil, fmt.Errorf("error getting document: %w", err)
	}

	return doc, nil
}

func (s *DocumentStore) GetDocumentsByType(docType string) ([]*DocumentEntry, error) {
	var docs []*DocumentEntry
	err := s.db.
		Where("document_type = ?", docType).
		Find(&docs).
		Error
	if err != nil {
		return nil, fmt.Errorf("error getting documents: %w", err)
	}

	return docs, nil
}

// GetDocumentsByRootID returns all documents with the given root id, not including
// the root document itself.
func (s *DocumentStore) GetDocumentsByRootID(id DocumentID) ([]*DocumentEntry, error) {
	var docs []*DocumentEntry
	err := s.db.
		Where("root_document_id = ?", id).
		Order("sequence ASC").
		Find(&docs).
		Error
	if err != nil {
		return nil, fmt.Errorf("error getting documents: %w", err)
	}

	return docs, nil
}
