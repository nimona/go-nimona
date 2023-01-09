package nimona

import (
	"fmt"
	"strings"

	"github.com/mr-tron/base58"
)

type DocumentID struct {
	_            string `cborgen:"$prefix,const=nimona://doc"`
	DocumentHash []byte
}

func (p DocumentID) String() string {
	return string(ResourceTypeDocumentID) + base58.Encode(p.DocumentHash)
}

func ParseDocumentID(pID string) (DocumentID, error) {
	prefix := string(ResourceTypeDocumentID)
	if !strings.HasPrefix(pID, prefix) {
		return DocumentID{}, fmt.Errorf("invalid resource id")
	}

	pID = strings.TrimPrefix(pID, prefix)
	hash, err := base58.Decode(pID)
	if err != nil {
		return DocumentID{}, fmt.Errorf("invalid resource id")
	}

	return DocumentID{DocumentHash: hash}, nil
}
