package nimona

import (
	"fmt"

	"nimona.io/internal/tilde"
)

type (
	DocumentPatch struct {
		_          string                   `nimona:"$type,type=core/stream/patch"`
		Metadata   Metadata                 `nimona:"$metadata,omitempty"`
		Operations []DocumentPatchOperation `nimona:"operations,omitempty"`
	}
	DocumentPatchOperation struct {
		Op    string      `nimona:"op"`
		Path  string      `nimona:"path"`
		Value tilde.Value `nimona:"value,omitempty"`
	}
)

func ApplyDocumentPatch(original *Document, docPatch *DocumentPatch) (*Document, error) {
	doc := original.Copy()
	docMap := doc.Map()
	for _, patch := range docPatch.Operations {
		switch patch.Op {
		case "replace":
			err := docMap.Set(patch.Path, patch.Value)
			if err != nil {
				return nil, fmt.Errorf("error applying patch: %w", err)
			}
		default:
			return nil, fmt.Errorf("unsupported patch operation: %s", patch.Op)
		}
	}
	return NewDocumentMap(docMap), nil
}
