package blocks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnpack(t *testing.T) {
	p := &Block{
		Type: "a",
		Payload: map[string]interface{}{
			"a": "a-value",
			"c": 12,
			"d": []byte{1, 2, 3},
		},
	}

	ev := &TestPackStructA{
		A: "a-value",
		C: 12,
		D: []byte{1, 2, 3},
	}

	v := &TestPackStructA{}
	err := UnpackInto(p, v)
	assert.NoError(t, err)
	assert.Equal(t, ev, v)
}

func TestUnpackNestedTyped(t *testing.T) {
	p := &Block{
		Type: "a",
		Payload: map[string]interface{}{
			"a": "a-value",
			"b": "HCsVC4MT3AwJMXPUoHfMgQSG6GbaMDMJxpueaBFFbD6YVqF7",
			"c": 12,
			"d": []byte{1, 2, 3},
		},
	}

	ev := &TestPackStructA{
		A: "a-value",
		B: &TestPackStructB{
			AA: "aa-value",
			BB: 1234,
		},
		C: 12,
		D: []byte{1, 2, 3},
	}

	v := &TestPackStructA{}
	err := UnpackInto(p, v)
	assert.NoError(t, err)
	assert.Equal(t, ev, v)
}

func TestUnpackDecodeNestedTyped(t *testing.T) {
	p := "2Jn7th59YMCscZggcVxGbpBVEboF4QdP2JBj9CkPJP3j7FqKC6736z6iHPFvWBop9kccPgyQ6523qXwZtcMk727eJhEhrXG4t54gkz3z4YaZYWwv6GALbcZRz"

	ev := &TestPackStructA{
		A: "a-value",
		B: &TestPackStructB{
			AA: "aa-value",
			BB: 1234,
		},
		C: 12,
		D: []byte{1, 2, 3},
	}

	v, err := UnpackDecodeBase58(p)
	assert.NoError(t, err)
	assert.Equal(t, ev, v)
}
