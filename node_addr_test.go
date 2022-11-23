package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeAddr(t *testing.T) {
	a := NewNodeAddr("utp", "localhost", 1234)
	require.Equal(t, "utp", a.Network())
	require.Equal(t, "localhost:1234", a.Address())
	require.Equal(t, "utp://localhost:1234", a.String())
}
