package hyperspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunk(t *testing.T) {
	b := []byte{0, 1, 2, 3, 4, 5, 6}
	r := chunk(b, 2)
	assert.Len(t, r, 4)
	assert.Equal(t, []byte{0, 1}, r[0])
	assert.Equal(t, []byte{2, 3}, r[1])
	assert.Equal(t, []byte{4, 5}, r[2])
	assert.Equal(t, []byte{6}, r[3])

	b = []byte{0, 1, 2, 3, 4, 5}
	r = chunk(b, 2)
	assert.Len(t, r, 3)
	assert.Equal(t, []byte{0, 1}, r[0])
	assert.Equal(t, []byte{2, 3}, r[1])
	assert.Equal(t, []byte{4, 5}, r[2])
}
