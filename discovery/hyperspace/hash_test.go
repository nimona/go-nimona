package hyperspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	v := "0"
	b := Hash([]byte(v))
	assert.Equal(t, int32(246), b)

	v = "c_0_0"
	b = Hash([]byte(v))
	assert.Equal(t, int32(203), b)
}
