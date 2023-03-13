package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/internal/tilde"
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
			Op:    "append",
			Path:  "strings",
			Value: tilde.String("foo"),
		}},
	}

	patch3 := &DocumentPatch{
		Metadata: Metadata{
			Root: &rootDocID,
		},
		Operations: []DocumentPatchOperation{{
			Op:    "append",
			Path:  "strings",
			Value: tilde.String("bar"),
		}},
	}

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

	applied, err := ApplyDocumentPatch(
		rootDoc,
		patch1,
		patch2,
		patch3,
	)
	require.NoError(t, err)
	require.Equal(t, exp.Map(), applied.Map())
}
