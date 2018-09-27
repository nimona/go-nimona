package primitives

import (
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/stretchr/testify/assert"
)

func TestBlockToMap(t *testing.T) {
	b := &Block{
		Type: "type",
		Annotations: &Annotations{
			Policies: []Policy{
				Policy{
					Subjects: []string{"recipient1"},
					Actions:  []string{"read"},
					Effect:   "allow",
				},
			},
		},
		Payload: map[string]interface{}{
			"foo": "bar",
		},
		Signature: &Signature{
			Key: &Key{
				Algorithm: "key-alg",
			},
			Alg: "sig-alg",
		},
	}

	em := map[string]interface{}{
		"type": "type",
		"annotations": map[string]interface{}{
			"policies": []interface{}{
				map[string]interface{}{
					"subjects": []string{
						"recipient1",
					},
					"actions": []string{
						"read",
					},
					"effect": "allow",
				},
			},
		},
		"payload": map[string]interface{}{
			"foo": "bar",
		},
		"signature": map[string]interface{}{
			"type": "nimona.io/signature",
			"payload": map[string]interface{}{
				"alg": "sig-alg",
				"key": map[string]interface{}{
					"type": "nimona.io/key",
					"payload": map[string]interface{}{
						"alg": "key-alg",
					},
				},
			},
		},
	}

	m := BlockToMap(b)
	spew.Dump(m)
	assert.Equal(t, em, m)
}

func TestBlockFromMap(t *testing.T) {
	eb := &Block{
		Type: "type",
		Payload: map[string]interface{}{
			"foo": "bar",
		},
		Signature: &Signature{
			Key: &Key{
				Algorithm: "key-alg",
			},
			Alg: "sig-alg",
		},
	}

	m := map[string]interface{}{
		"type": "type",
		"payload": map[string]interface{}{
			"foo": "bar",
		},
		"signature": map[string]interface{}{
			"type": "nimona.io/signature",
			"payload": map[string]interface{}{
				"alg": "sig-alg",
				"key": map[string]interface{}{
					"type": "nimona.io/key",
					"payload": map[string]interface{}{
						"alg": "key-alg",
					},
				},
			},
		},
	}

	b := BlockFromMap(m)
	spew.Dump(b)
	assert.Equal(t, eb, b)
}
