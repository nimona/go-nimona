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

	o := FromMap(m)

	assert.Equal(t, m["@ctx:s"], o.GetRaw("@ctx:s"))
	assert.Equal(t, m["@signature:o"], o.GetRaw("@signature:o"))
	assert.Equal(t, m["@policy:o"], o.GetRaw("@policy:o"))
	assert.Equal(t, m["@parents:as"], o.GetRaw("@parents:as"))

	n := New()

	n.SetRaw("@ctx:", m["@ctx:s"])
	n.SetRaw("@signature:o", m["@signature:o"])
	n.SetRaw("@policy:o", m["@policy:o"])
	n.SetRaw("@parents:as", m["@parents:as"])

	assert.Equal(t, m["@ctx:s"], n.GetRaw("@ctx:"))
	assert.Equal(t, m["@signature:o"], n.GetRaw("@signature:o"))
	assert.Equal(t, m["@policy:o"], n.GetRaw("@policy:o"))
	assert.Equal(t, m["@parents:as"], n.GetRaw("@parents:as"))

	e := New()

	e.SetType(o.GetType())
	s := o.GetSignature()
	e.SetSignature(*s)
	p := o.GetPolicy()
	e.SetPolicy(*p)
	e.SetParents(o.GetParents())

	assert.NotNil(t, e.GetRaw("@ctx:s"))
	assert.NotNil(t, e.GetRaw("@signature:o"))
	assert.NotNil(t, e.GetRaw("@policy:o"))
	assert.NotNil(t, e.GetRaw("@parents:as"))

	assert.Equal(t, m["@ctx:s"], e.GetRaw("@ctx:s"))
	assert.Equal(t, m["@signature:o"], e.GetRaw("@signature:o"))
	assert.Equal(t, m["@policy:o"], e.GetRaw("@policy:o"))
	assert.Equal(t, m["@parents:as"], e.GetRaw("@parents:as"))
}
