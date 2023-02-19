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
		Value DocumentMap `nimona:"value,omitempty"`
	}
)

