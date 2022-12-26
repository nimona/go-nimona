package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNetworkID(t *testing.T) {
	s0 := "nimona://network:handle:testing.reamde.dev"
	n0 := NetworkID{
		Hostname: "testing.reamde.dev",
	}

	require.Equal(t, s0, n0.String())

	n1, err := ParseNetworkID(s0)
	require.NoError(t, err)
	require.Equal(t, n0, n1)
}
