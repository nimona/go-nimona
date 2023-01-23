package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDocumentType(t *testing.T) {
	require.Equal(t, "nimona://peer:addr:", DocumentTypePeerAddress.String())
}
