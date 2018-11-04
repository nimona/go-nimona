package hyperspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	v := "foo"
	b := Hash([]byte(v))
	assert.Equal(t, int32(289), b)
}
