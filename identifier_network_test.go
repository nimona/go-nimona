package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNetworkID_ParseNetworkID(t *testing.T) {
	s0 := "nimona://network:handle:testing.reamde.dev"
	n0 := NetworkID{
		Hostname: "testing.reamde.dev",
	}

	require.Equal(t, s0, n0.String())

	n1, err := ParseNetworkID(s0)
	require.NoError(t, err)
	require.Equal(t, n0, n1)
}

func TestNetworkID_MarshalUnmarshal(t *testing.T) {
	n0 := NetworkID{
		Hostname: "testing.reamde.dev",
	}

	b, err := n0.MarshalCBORBytes()
	require.NoError(t, err)

	var n1 NetworkID
	err = n1.UnmarshalCBORBytes(b)
	require.NoError(t, err)

	require.Equal(t, n0, n1)
}
