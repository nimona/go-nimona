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

	patch := &DocumentPatch{
		Metadata: Metadata{
			Root: &rootDocID,
		},
		Operations: []DocumentPatchOperation{{
			Op:    "replace",
			Path:  "displayName",
			Value: tilde.String("bar"),
		}},
	}

	exp := NewDocumentMap(
		tilde.Map{
			"$type":       tilde.String("core/identity/profile"),
			"displayName": tilde.String("bar"),
		},
	)

	applied, err := ApplyDocumentPatch(
		rootDoc,
		patch,
	)
	require.NoError(t, err)
	require.Equal(t, exp.Map(), applied.Map())
}
