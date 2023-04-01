package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ParseDocumentNRI(t *testing.T) {
	id := NewTestRandomDocumentID(t)
	got, err := ParseDocumentNRI(id.String())
	require.NoError(t, err)
	require.Equal(t, id, got)
}
