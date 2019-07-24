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
		"@parents:as": []string{"parent-value"},
	}

	em := map[string]interface{}{
		"@ctx:s": "ctx-value",
		"@signature:o": map[string]interface{}{
			"@ctx:s": "-signature",
		},
		"@policy:o": map[string]interface{}{
			"@ctx:s": "-policy",
		},
		"@parents:as": []string{"parent-value"},
	}

	o := FromMap(m)

	assert.Equal(t, m["@ctx:s"], o.Get("@ctx:s"))
	assert.Equal(t, m["@signature:o"], o.Get("@signature:o"))
	assert.Equal(t, m["@policy:o"], o.Get("@policy:o"))
	assert.Equal(t, m["@parents:as"], o.Get("@parents:as"))

	n := New()

	n.Set("@ctx:", m["@ctx:s"])
	n.Set("@signature:o", m["@signature:o"])
	n.Set("@policy:o", m["@policy:o"])
	n.Set("@parents:as", m["@parents:as"])

	assert.Equal(t, em["@ctx:s"], n.Get("@ctx:"))
	assert.Equal(t, em["@signature:o"], n.Get("@signature:o"))
	assert.Equal(t, em["@policy:o"], n.Get("@policy:o"))
	assert.Equal(t, em["@parents:as"], n.Get("@parents:as"))

	e := New()

	e.SetType(o.GetType())
	s := o.GetSignature()
	e.SetSignature(*s)
	p := o.GetPolicy()
	e.SetPolicy(*p)
	e.SetParents(o.GetParents())

	assert.NotNil(t, e.Get("@ctx:s"))
	assert.NotNil(t, e.Get("@signature:o"))
	assert.NotNil(t, e.Get("@policy:o"))
	assert.NotNil(t, e.Get("@parents:as"))

	assert.Equal(t, em["@ctx:s"], e.Get("@ctx:s"))
	assert.Equal(t, em["@signature:o"], e.Get("@signature:o"))
	assert.Equal(t, em["@policy:o"], e.Get("@policy:o"))
	assert.Equal(t, em["@parents:as"], e.Get("@parents:as"))
}
