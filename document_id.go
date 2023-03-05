package nimona

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type (
	DocumentID struct {
		_            string       `nimona:"$type,type=core/document/id"`
		DocumentHash DocumentHash `nimona:"hash,omitempty"`
	}
)

func NewDocumentID(m *Document) DocumentID {
	return DocumentID{
		DocumentHash: NewDocumentHash(m),
	}
}

func ParseDocumentID(pID string) (DocumentID, error) {
	prefix := string(ShorthandDocumentID)
	if !strings.HasPrefix(pID, prefix) {
		return DocumentID{}, fmt.Errorf("invalid resource id")
	}

	pID = strings.TrimPrefix(pID, prefix)
	hash, err := ParseDocumentHash(pID)
	if err != nil {
		return DocumentID{}, fmt.Errorf("invalid resource id")
	}

	return DocumentID{DocumentHash: hash}, nil
}

func (p DocumentID) String() string {
	return string(ShorthandDocumentID) + p.DocumentHash.String()
}

func (p DocumentID) IsEmpty() bool {
	return len(p.DocumentHash) == 0
}

func (p DocumentID) IsEqual(other DocumentID) bool {
	return p.DocumentHash.IsEqual(other.DocumentHash)
}

func (p DocumentID) Value() (driver.Value, error) {
	return p.String(), nil
}

func (p *DocumentID) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if docIDString, ok := value.(string); ok {
		docID, err := ParseDocumentID(docIDString)
		if err != nil {
			return fmt.Errorf("unable to scan into DocumentID: %w", err)
		}
		p.DocumentHash = docID.DocumentHash
		return nil
	}
	return fmt.Errorf("unable to scan %T into DocumentID", value)
}
