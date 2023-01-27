package nimona

import "fmt"

func GetDocumentTypeFromCbor(cborBytes []byte) (string, error) {
	doc := &DocumentBase{}
	err := UnmarshalCBORBytes(cborBytes, doc)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling document: %w", err)
	}
	return doc.Type, nil
}
