package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectMethods(t *testing.T) {
	m := map[string]interface{}{
		"@ctx:s": "ctx-value",
		"@signature:o": Object{
			"@ctx:s": "-signature",
		},
		"@policy:o": Object{
			"@ctx:s": "-policy",
		},
		"@parents:ao": []Object{
			(&Hash{}).ToObject(),
		},
	}

	em := map[string]interface{}{
		"@ctx:s": "ctx-value",
		"@signature:o": map[string]interface{}{
			"@ctx:s": "-signature",
		},
		"@policy:o": map[string]interface{}{
			"@ctx:s": "-policy",
		},
		"@parents:ao": []map[string]interface{}{
			map[string]interface{}{
				"@ctx:s":    "nimona.io/Hash",
				"@domain:s": "nimona.io/object",
				"@struct:s": "Hash",
			},
		},
	}

	o := FromMap(m)

	assert.Equal(t, em["@ctx:s"], o.Get("@ctx:s"))
	assert.Equal(t, em["@signature:o"], o.Get("@signature:o"))
	assert.Equal(t, em["@policy:o"], o.Get("@policy:o"))

	n := New()

	n.Set("@ctx:", m["@ctx:s"])
	n.Set("@signature:o", m["@signature:o"])
	n.Set("@policy:o", m["@policy:o"])

	assert.Equal(t, em["@ctx:s"], n.Get("@ctx:"))
	assert.Equal(t, em["@signature:o"], n.Get("@signature:o"))
	assert.Equal(t, em["@policy:o"], n.Get("@policy:o"))

	e := New()

	e.SetType(o.GetType())
	s := o.GetSignature()
	e.SetSignature(*s)

	assert.NotNil(t, e.Get("@ctx:s"))
	assert.NotNil(t, e.Get("@signature:o"))

	assert.Equal(t, em["@ctx:s"], e.Get("@ctx:s"))
	assert.Equal(t, em["@signature:o"], e.Get("@signature:o"))
}
