package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentifiers(t *testing.T) {
	require.Equal(t, "nimona://peer:addr:", ResourceTypePeerAddress.String())
}
