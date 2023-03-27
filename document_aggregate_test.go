package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"

	"nimona.io/internal/ckv"
	"nimona.io/tilde"
)

func TestDocumentAggregate_Apply(t *testing.T) {
	// db, err := gorm.Open(
	// 	sqlite.Open("./test.db?cache=shared"),
	// 	&gorm.Config{},
	// )
	// require.NoError(t, err)

	db := NewTestDocumentDB(t)
	store := ckv.NewSQLStore(db)
	aggregate := &AggregateStore{
		Store: store,
	}

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

	patch4doc := &ProfileRepository{
		Alias:  "testing.nimona.dev",
		Handle: "foo",
	}

	patch4 := &DocumentPatch{
		Metadata: Metadata{
			Root: &rootDocID,
		},
		Operations: []DocumentPatchOperation{{
			Op:        "append",
			Path:      "repositories",
			Value:     patch4doc.Map(),
			Key:       "9e81fff2-ec7a-44e2-aa05-617c84bbaf5e",
			Partition: []string{"2023", "04"},
		}},
	}

	aggregate.Apply(rootDoc)
	aggregate.Apply(patch1.Document())
	aggregate.Apply(patch4.Document())

	res := []*ProfileRepository{}
	err := aggregate.GetAggregateNested(
		rootDocID,
		"repositories",
		&res,
	)
	require.NoError(t, err)
	require.Equal(t, 1, len(res))
	require.Equal(t, patch4doc, res[0])
}
