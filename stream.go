package nimona

import (
	"encoding/json"
	"fmt"

	cborpatch "github.com/ldclabs/cbor-patch"
	"github.com/wI2L/jsondiff"
	cbg "github.com/whyrusleeping/cbor-gen"
)

type StreamID struct {
	Hash string
}

type StreamInfo struct {
	StreamID   StreamID
	Operations []StreamOperation
}

type StreamOperation struct {
	Op    string       `cborgen:"op"`
	Path  string       `cborgen:"path"`
	From  string       `cborgen:"from,omitempty"`
	Value cbg.Deferred `cborgen:"value,omitempty"`
}

type StreamPatch struct {
	_            string            `cborgen:"$type,const=core/stream/patch"`
	Dependencies []DocumentID      `cborgen:"dependencies,omitempty"`
	Operations   []StreamOperation `cborgen:"operations,omitempty"`
}

func streamPatchFromCBORPatch(cp cborpatch.Patch) StreamPatch {
	ops := make([]StreamOperation, len(cp))
	for i, op := range cp {
		ops[i] = StreamOperation{
			Op:    op.Op,
			Path:  op.Path,
			From:  op.From,
			Value: cbg.Deferred{Raw: op.Value},
		}
	}
	return StreamPatch{
		Operations: ops,
	}
}

func streamPatchToCBORPatch(p StreamPatch) cborpatch.Patch {
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

func CreateStreamPatch(
	original Cborer,
	updated Cborer,
) (StreamPatch, error) {
	p, err := createCBORPatch(original, updated)
	if err != nil {
		return StreamPatch{}, fmt.Errorf("error creating json patch: %w", err)
	}

	return streamPatchFromCBORPatch(p), nil
}

func ApplyStreamPatch(
	original Cborer,
	patches ...StreamPatch,
) error {
	oc, err := original.MarshalCBORBytes()
	if err != nil {
		return fmt.Errorf("error marshaling original: %w", err)
	}

	for _, sp := range patches {
		p := streamPatchToCBORPatch(sp)
		rc, err := p.Apply(oc)
		if err != nil {
			return fmt.Errorf("error applying patch: %w", err)
		}
		err = original.UnmarshalCBORBytes(rc)
		if err != nil {
			return fmt.Errorf("error unmarshaling patched: %w", err)
		}
	}

	return nil
}

func createCBORPatch(
	original Cborer,
	updated Cborer,
) (cborpatch.Patch, error) {
	d, err := jsondiff.Compare(original, updated)
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
