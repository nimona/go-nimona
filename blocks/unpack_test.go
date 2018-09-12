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
			"b": map[string]interface{}{
				"payload": map[string]interface{}{
					"aa": "aa-value",
					"bb": 1234,
				},
				"type": "b",
			},
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
	p := "BafYRz7KhUdXVegjc94yBfMew2hJZPT2mPu2pYwGXAYLL7VXik45E9b3MbtFLLDTyYMcjcvoHPVbL61eD8hP5uyxXDN9ssEqZSdN"

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
