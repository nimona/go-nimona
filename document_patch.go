package nimona

import (
	"fmt"

	cborpatch "github.com/ldclabs/cbor-patch"
)

type (
	DocumentPatch struct {
		_            string                   `nimona:"$type,type=core/stream/patch"`
		Metadata     Metadata                 `nimona:"$metadata,omitempty"`
		Dependencies []DocumentID             `nimona:"dependencies,omitempty"`
		Operations   []DocumentPatchOperation `nimona:"operations,omitempty"`
	}
	DocumentPatchOperation struct {
		Op    string      `nimona:"op"`
		Path  string      `nimona:"path"`
		From  string      `nimona:"from,omitempty"`
		Value DocumentMap `nimona:"value,omitempty"`
	}
)

// func documentPatchFromCBORPatch(cp cborpatch.Patch) *DocumentPatch {
// 	ops := make([]DocumentPatchOperation, len(cp))
// 	for i, op := range cp {
// 		ops[i] = DocumentPatchOperation{
// 			Op:    op.Op,
// 			Path:  op.Path,
// 			From:  op.From,
// 			Value: cbg.Deferred{Raw: op.Value},
// 		}
// 	}
// 	return &DocumentPatch{
// 		Operations: ops,
// 	}
// }

func documentPatchToCBORPatch(p *DocumentPatch) cborpatch.Patch {
	ops := make([]cborpatch.Operation, len(p.Operations))
	for i, op := range p.Operations {
		// TODO(geaoh): fix this
		cb, _ := op.Value.MarshalCBOR()
		ops[i] = cborpatch.Operation{
			Op:    op.Op,
			Path:  op.Path,
			From:  op.From,
			Value: cb,
		}
	}
	return ops
}

// func CreateDocumentPatch(
// 	originalCbor []byte,
// 	updatedCbor []byte,
// ) (*DocumentPatch, error) {
// 	p, err := createCBORPatch(originalCbor, updatedCbor)
// 	if err != nil {
// 		return nil, fmt.Errorf("error creating json patch: %w", err)
// 	}

// 	return documentPatchFromCBORPatch(p), nil
// }

func ApplyDocumentPatch(
	original DocumentMapper,
	patches ...*DocumentPatch,
) error {
	oc, err := original.DocumentMap().MarshalCBOR()
	if err != nil {
		return fmt.Errorf("error marshaling original: %w", err)
	}

	for _, sp := range patches {
		p := documentPatchToCBORPatch(sp)
		rc, err := p.Apply(oc)
		if err != nil {
			return fmt.Errorf("error applying patch: %w", err)
		}
		err = UnmarshalCBORBytes(rc, original)
		if err != nil {
			return fmt.Errorf("error unmarshaling patched: %w", err)
		}
	}

	return nil
}

// func createCBORPatch(
// 	originalCbor []byte,
// 	updatedCbor []byte,
// ) (cborpatch.Patch, error) {
// 	originalMap, err := NewDocumentMapFromCBOR(originalCbor)
// 	if err != nil {
// 		return nil, fmt.Errorf("error converting original cbor to map: %w", err)
// 	}

// 	updatedMap, err := NewDocumentMapFromCBOR(updatedCbor)
// 	if err != nil {
// 		return nil, fmt.Errorf("error converting updated cbor to map: %w", err)
// 	}

// 	d, err := jsondiff.Compare(originalMap, updatedMap)
// 	if err != nil {
// 		return nil, fmt.Errorf("error comparing json: %w", err)
// 	}

// 	dj, err := json.Marshal(d)
// 	if err != nil {
// 		return nil, fmt.Errorf("error marshaling diff: %w", err)
// 	}

// 	pc, err := cborpatch.FromJSON(dj, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("error decoding patch: %w", err)
// 	}

// 	p, err := cborpatch.NewPatch(pc)
// 	if err != nil {
// 		return nil, fmt.Errorf("error decoding patch: %w", err)
// 	}

// 	return p, nil
// }
