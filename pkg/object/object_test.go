package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectMethods(t *testing.T) {
	m := map[string]interface{}{
		"@type:s": "ctx-value",
		"@signature:o": Object{
			"@type:s": "-signature",
		},
		"something:o": Object{
			"@type:s": "-something",
		},
		"parents:ao": []Object{
			(&Hash{}).ToObject(),
		},
	}

	em := map[string]interface{}{
		"@type:s": "ctx-value",
		"@signature:o": map[string]interface{}{
			"@type:s": "-signature",
		},
		"something:o": map[string]interface{}{
			"@type:s": "-something",
		},
		"parents:ao": []map[string]interface{}{
			map[string]interface{}{
				"@type:s":   "nimona.io/Hash",
				"@domain:s": "nimona.io/object",
				"@struct:s": "Hash",
			},
		},
	}

	o := FromMap(m)

	assert.Equal(t, em["@type:s"], o.Get("@type:s"))
	assert.Equal(t, em["@signature:o"], o.Get("@signature:o"))
	assert.Equal(t, em["something:o"], o.Get("something:o"))

	n := New()

	n.Set("@type:", m["@type:s"])
	n.Set("@signature:o", m["@signature:o"])
	n.Set("something:o", m["something:o"])

	assert.Equal(t, em["@type:s"], n.Get("@type:"))
	assert.Equal(t, em["@signature:o"], n.Get("@signature:o"))
	assert.Equal(t, em["something:o"], n.Get("something:o"))

	e := New()

	e.SetType(o.GetType())
	s := o.GetSignature()
	e.SetSignature(*s)

	assert.NotNil(t, e.Get("@type:s"))
	assert.NotNil(t, e.Get("@signature:o"))

	assert.Equal(t, em["@type:s"], e.Get("@type:s"))
	assert.Equal(t, em["@signature:o"], e.Get("@signature:o"))
}
