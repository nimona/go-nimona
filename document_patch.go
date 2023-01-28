package nimona

import (
	"encoding/json"
	"fmt"

	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/wI2L/jsondiff"
	cbg "github.com/whyrusleeping/cbor-gen"
)

type (
	DocumentPatch struct {
		_            string                   `cborgen:"$type,const=core/stream/patch"`
		Metadata     Metadata                 `cborgen:"$metadata,omitempty"`
		Dependencies []DocumentID             `cborgen:"dependencies,omitempty"`
		Operations   []DocumentPatchOperation `cborgen:"operations,omitempty"`
	}
	DocumentPatchOperation struct {
		Op    string       `cborgen:"op"`
		Path  string       `cborgen:"path"`
		From  string       `cborgen:"from,omitempty"`
		Value cbg.Deferred `cborgen:"value,omitempty"`
	}
)

func documentPatchFromCBORPatch(cp cborpatch.Patch) *DocumentPatch {
	ops := make([]DocumentPatchOperation, len(cp))
	for i, op := range cp {
		ops[i] = DocumentPatchOperation{
			Op:    op.Op,
			Path:  op.Path,
			From:  op.From,
			Value: cbg.Deferred{Raw: op.Value},
		}
	}
	return &DocumentPatch{
		Operations: ops,
	}
}

func documentPatchToCBORPatch(p *DocumentPatch) cborpatch.Patch {
	ops := make([]cborpatch.Operation, len(p.Operations))
	for i, op := range p.Operations {
		ops[i] = cborpatch.Operation{
			Op:    op.Op,
			Path:  op.Path,
			From:  op.From,
			Value: op.Value.Raw,
		}
	}
	return ops
}

func CreateDocumentPatch(
	originalCbor []byte,
	updatedCbor []byte,
) (*DocumentPatch, error) {
	p, err := createCBORPatch(originalCbor, updatedCbor)
	if err != nil {
		return nil, fmt.Errorf("error creating json patch: %w", err)
	}

	return documentPatchFromCBORPatch(p), nil
}

func ApplyDocumentPatch(
	original Cborer,
	patches ...*DocumentPatch,
) error {
	oc, err := MarshalCBORBytes(original)
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

func createCBORPatch(
	originalCbor []byte,
	updatedCbor []byte,
) (cborpatch.Patch, error) {
	originalMap, err := NewDocumentMapFromCBOR(originalCbor)
	if err != nil {
		return nil, fmt.Errorf("error converting original cbor to map: %w", err)
	}

	updatedMap, err := NewDocumentMapFromCBOR(updatedCbor)
	if err != nil {
		return nil, fmt.Errorf("error converting updated cbor to map: %w", err)
	}

	d, err := jsondiff.Compare(originalMap, updatedMap)
	if err != nil {
		return nil, fmt.Errorf("error comparing json: %w", err)
	}

	dj, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("error marshaling diff: %w", err)
	}

	pc, err := cborpatch.FromJSON(dj, nil)
	if err != nil {
		return nil, fmt.Errorf("error decoding patch: %w", err)
	}

	p, err := cborpatch.NewPatch(pc)
	if err != nil {
		return nil, fmt.Errorf("error decoding patch: %w", err)
	}

	return p, nil
}
