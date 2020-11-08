package hyperspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBloom(t *testing.T) {
	b := New("a", "b", "c")
	assert.True(t, b.Test(New("a")))
	assert.True(t, b.Test(New("b")))
	assert.True(t, b.Test(New("c")))
	assert.False(t, b.Test(New("d")))
	assert.False(t, b.Test(New("e")))
	assert.False(t, b.Test(New("f")))
}
