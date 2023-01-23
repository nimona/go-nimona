package nimona

import "fmt"

func GetDocumentTypeFromCbor(cborBytes []byte) (string, error) {
	doc := &DocumentBase{}
	err := doc.UnmarshalCBORBytes(cborBytes)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling document: %w", err)
	}
	return doc.Type, nil
}
