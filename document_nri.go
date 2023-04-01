package nimona

import (
	"fmt"
	"strings"
)

func ParseDocumentNRI(nri string) (DocumentID, error) {
	prefix := string(ShorthandDocumentID)
	if !strings.HasPrefix(nri, prefix) {
		return DocumentID{}, fmt.Errorf("invalid resource id")
	}

	nri = strings.TrimPrefix(nri, prefix)
	hash, err := ParseDocumentHash(nri)
	if err != nil {
		return DocumentID{}, fmt.Errorf("invalid resource id")
	}

	return DocumentID{DocumentHash: hash}, nil
}
