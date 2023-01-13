package nimona

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

func NewDocumentID(v Cborer) DocumentID {
	hash, err := NewDocumentHash(v)
	if err != nil {
		panic(fmt.Errorf("error creating document id: %w", err))
	}
	return DocumentID{
		DocumentHash: hash,
	}
}

func NewDocumentIDFromCBOR(b []byte) DocumentID {
	hash, err := NewDocumentHashFromCBOR(b)
	if err != nil {
		panic(fmt.Errorf("error creating document id: %w", err))
	}
	return DocumentID{
		DocumentHash: hash,
	}
}

func ParseDocumentID(pID string) (DocumentID, error) {
	prefix := string(ResourceTypeDocumentID)
	if !strings.HasPrefix(pID, prefix) {
		return DocumentID{}, fmt.Errorf("invalid resource id")
	}

	pID = strings.TrimPrefix(pID, prefix)
	hash, err := DocumentHashFromBase58(pID)
	if err != nil {
		return DocumentID{}, fmt.Errorf("invalid resource id")
	}

	return DocumentID{DocumentHash: hash}, nil
}

type DocumentID struct {
	_            string `cborgen:"$prefix,const=nimona://doc"`
	DocumentHash DocumentHash
}

func (p DocumentID) String() string {
	return string(ResourceTypeDocumentID) + p.DocumentHash.String()
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
