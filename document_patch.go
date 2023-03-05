package nimona

type (
	DocumentPatch struct {
		_            string                   `nimona:"$type,type=core/stream/patch"`
		Metadata     Metadata                 `nimona:"$metadata,omitempty"`
		Dependencies []DocumentID             `nimona:"dependencies,omitempty"`
		Operations   []DocumentPatchOperation `nimona:"operations,omitempty"`
	}
	DocumentPatchOperation struct {
		Op   string `nimona:"op"`
		Path string `nimona:"path"`
		From string `nimona:"from,omitempty"`
		// TODO: This should probably be a tilde.Map
		Value Document `nimona:"value,omitempty"`
	}
)

// func CreateDocumentPatch(original, target *Document) (*DocumentPatch, error) {
// 	jsonOps, err := jsondiff.Compare(original, target)
// 	if err != nil {
// 		return nil, fmt.Errorf("error creating patch: %w", err)
// 	}

// 	ops := []DocumentPatchOperation{}
// 	for _, op := range jsonOps {
// 		valJSON, err := json.Marshal(op)
// 		if err != nil {
// 			return nil, fmt.Errorf("error marshaling value: %w", err)
// 		}
// 		fmt.Println(string(valJSON))
// 		valMap := &Document{}
// 		err = valMap.UnmarshalJSON(valJSON)
// 		if err != nil {
// 			return nil, fmt.Errorf("error unmarshaling value: %w", err)
// 		}
// 		ops = append(ops, DocumentPatchOperation{
// 			Op:    op.Type,
// 			Path:  op.Path.String(),
// 			From:  op.From.String(),
// 			Value: *valMap,
// 		})
// 	}

// 	patch := &DocumentPatch{
// 		Operations: ops,
// 	}

// 	return patch, nil
// }

// func ApplyDocumentPatch(original *Document, docPatch *DocumentPatch) (*Document, error) {
// 	patchJSON, err := json.Marshal(docPatch.Operations)
// 	if err != nil {
// 		return nil, fmt.Errorf("error marshaling patch: %w", err)
// 	}

// 	originalJSON, err := json.Marshal(original)
// 	if err != nil {
// 		return nil, fmt.Errorf("error marshaling original: %w", err)
// 	}

// 	patch, err := jsonpatch.DecodePatch(patchJSON)
// 	if err != nil {
// 		return nil, fmt.Errorf("error decoding patch: %w", err)
// 	}

// 	finalJSON, err := patch.Apply(originalJSON)
// 	if err != nil {
// 		return nil, fmt.Errorf("error applying patch: %w", err)
// 	}

// 	final := &Document{}
// 	err = json.Unmarshal(finalJSON, final)
// 	if err != nil {
// 		return nil, fmt.Errorf("error unmarshaling final: %w", err)
// 	}

// 	return final, nil
// }
