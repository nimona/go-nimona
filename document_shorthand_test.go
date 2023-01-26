package nimona

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShorthand(t *testing.T) {
	require.Equal(t, "nimona://peer:addr:", ShorthandPeerAddress.String())
}
