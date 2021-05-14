package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshal(t *testing.T) {
	o := &Object{
		Data: Map{
			"string:s": String("string"),
		},
	}
	s := &TestMap{}
	err := Unmarshal(o, s)
	assert.NoError(t, err)
	assert.Equal(t, "string", s.String)
}
