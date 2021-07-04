package multiheader

import (
	"testing"

	"github.com/multiformats/go-multicodec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiheader_EncodeDecode(t *testing.T) {
	c := multicodec.P384Pub
	b := []byte{1, 2, 3}
	eb := Encode(c, b)
	gc, gb, err := Decode(eb)
	require.NoError(t, err)
	assert.Equal(t, b, gb)
	assert.Equal(t, c, gc)
}
