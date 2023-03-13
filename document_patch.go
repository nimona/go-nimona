package nimona

import (
	"fmt"

	"nimona.io/tilde"
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

func ApplyDocumentPatch(
	original *Document,
	patches ...*DocumentPatch,
) (*Document, error) {
	doc := original.Copy()
	docMap := doc.Map()
	for _, patch := range patches {
		for _, operation := range patch.Operations {
			switch operation.Op {
			case "replace":
				err := docMap.Set(operation.Path, operation.Value)
				if err != nil {
					return nil, fmt.Errorf("error applying patch: %w", err)
				}
			case "append":
				err := docMap.Append(operation.Path, operation.Value)
				if err != nil {
					return nil, fmt.Errorf("error applying patch: %w", err)
				}
			default:
				return nil, fmt.Errorf("unsupported operation: %s", operation.Op)
			}
		}
	}
	return NewDocument(docMap), nil
}
