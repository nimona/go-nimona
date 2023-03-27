package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/tilde"
)

func TestDocumentPatch_Apply(t *testing.T) {
	root := &Profile{
		DisplayName: "foo",
	}

	rootDoc := root.Document()
	rootDocID := NewDocumentID(rootDoc)

	patch1 := &DocumentPatch{
		Metadata: Metadata{
			Root: &rootDocID,
		},
		Operations: []DocumentPatchOperation{{
			Op:    "replace",
			Path:  "displayName",
			Value: tilde.String("bar"),
		}},
	}

	patch2 := &DocumentPatch{
		Metadata: Metadata{
			Root: &rootDocID,
		},
		Operations: []DocumentPatchOperation{{
			Op:        "append",
			Path:      "strings",
			Value:     tilde.String("foo"),
			Key:       "a6ed7f26-41b3-4cc8-9288-6450f5f8d2ae",
			Partition: []string{"2023", "03"},
		}},
	}

	patch3 := &DocumentPatch{
		Metadata: Metadata{
			Root: &rootDocID,
		},
		Operations: []DocumentPatchOperation{{
			Op:        "append",
			Path:      "strings",
			Value:     tilde.String("bar"),
			Key:       "9e81fff2-ec7a-44e2-aa05-617c84bbaf5e",
			Partition: []string{"2023", "04"},
		}},
	}

	// patch4doc := &ProfileRepository{
	// 	Alias:  "testing.nimona.dev",
	// 	Handle: "foo",
	// }
	// patch4 := &DocumentPatch{
	// 	Metadata: Metadata{
	// 		Root: &rootDocID,
	// 	},
	// 	Operations: []DocumentPatchOperation{{
	// 		Op:        "append",
	// 		Path:      "repositories",
	// 		Value:     patch4doc.Map(),
	// 		Key:       "9e81fff2-ec7a-44e2-aa05-617c84bbaf5e",
	// 		Partition: []string{"2023", "04"},
	// 	}},
	// }

	exp := NewDocument(
		tilde.Map{
			"$type":       tilde.String("core/identity/profile"),
			"displayName": tilde.String("bar"),
			"strings": tilde.List{
				tilde.String("foo"),
				tilde.String("bar"),
			},
		},
	)

	DumpDocument(rootDoc)
	DumpDocument(patch1.Document())
	DumpDocument(patch2.Document())
	DumpDocument(patch3.Document())

	applied, err := ApplyDocumentPatch(
		rootDoc,
		patch1,
		patch2,
		patch3,
	)
	require.NoError(t, err)
	require.Equal(t, exp.Map(), applied.Map())
}
