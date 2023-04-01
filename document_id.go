package nimona

import (
	"database/sql/driver"
	"fmt"
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

func (p DocumentID) String() string {
	if p.IsEmpty() {
		return ""
	}
	return string(ShorthandDocumentID) + p.DocumentHash.String()
}

func (p DocumentID) IsEmpty() bool {
	return p.DocumentHash.IsEmpty()
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
		docID, err := ParseDocumentNRI(docIDString)
		if err != nil {
			return fmt.Errorf("unable to scan into DocumentID: %w", err)
		}
		p.DocumentHash = docID.DocumentHash
		return nil
	}
	return fmt.Errorf("unable to scan %T into DocumentID", value)
}
