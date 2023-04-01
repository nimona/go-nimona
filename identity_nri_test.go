package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentity_ParseIdentityNRI(t *testing.T) {
	id := NewTestIdentity(t)
	require.Equal(t, id.String(), id.String())

	id, err := ParseIdentityNRI(id.String())
	require.NoError(t, err)
	require.Equal(t, id.String(), id.String())
}
