package object

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectMethods(t *testing.T) {
	m := map[string]interface{}{
		"@ctx:s": "ctx-value",
		"@signature:o": map[string]interface{}{
			"@ctx:s": "-signature",
		},
		"@authority:o": map[string]interface{}{
			"@ctx:s": "-authority",
		},
		"@signer:o": map[string]interface{}{
			"@ctx:s": "-signer",
		},
		"@policy:o": map[string]interface{}{
			"@ctx:s": "-policy",
		},
		"@mandate:o": map[string]interface{}{
			"@ctx:s": "-mandate",
		},
		"@parents:a<s>": []string{"parent-value"},
	}

	o := FromMap(m)

	assert.Equal(t, jp(m["@ctx:s"]), jp(o.GetRaw("@ctx")))
	assert.Equal(t, jp(m["@signature:o"]), jp(o.GetRaw("@signature")))
	assert.Equal(t, jp(m["@authority:o"]), jp(o.GetRaw("@authority")))
	assert.Equal(t, jp(m["@signer:o"]), jp(o.GetRaw("@signer")))
	assert.Equal(t, jp(m["@policy:o"]), jp(o.GetRaw("@policy")))
	assert.Equal(t, jp(m["@mandate:o"]), jp(o.GetRaw("@mandate")))
	assert.Equal(t, jp(m["@parents:a<s>"]), jp(o.GetRaw("@parents")))

	n := New()

	n.SetRaw("@ctx", m["@ctx:s"])
	n.SetRaw("@signature", m["@signature:o"])
	n.SetRaw("@authority", m["@authority:o"])
	n.SetRaw("@signer", m["@signer:o"])
	n.SetRaw("@policy", m["@policy:o"])
	n.SetRaw("@mandate", m["@mandate:o"])
	n.SetRaw("@parents", m["@parents:a<s>"])

	assert.Equal(t, jp(m["@ctx:s"]), jp(n.GetRaw("@ctx")))
	assert.Equal(t, jp(m["@signature:o"]), jp(n.GetRaw("@signature")))
	assert.Equal(t, jp(m["@authority:o"]), jp(n.GetRaw("@authority")))
	assert.Equal(t, jp(m["@signer:o"]), jp(n.GetRaw("@signer")))
	assert.Equal(t, jp(m["@policy:o"]), jp(n.GetRaw("@policy")))
	assert.Equal(t, jp(m["@mandate:o"]), jp(n.GetRaw("@mandate")))
	assert.Equal(t, jp(m["@parents:a<s>"]), jp(n.GetRaw("@parents")))

	e := New()

	e.SetType(o.GetType())
	e.SetSignature(o.GetSignature())
	e.SetAuthorityKey(o.GetAuthorityKey())
	e.SetSignerKey(o.GetSignerKey())
	e.SetPolicy(o.GetPolicy())
	e.SetMandate(o.GetMandate())
	e.SetParents(o.GetParents())

	assert.NotNil(t, e.GetRaw("@ctx"))
	assert.NotNil(t, e.GetRaw("@signature"))
	assert.NotNil(t, e.GetRaw("@authority"))
	assert.NotNil(t, e.GetRaw("@signer"))
	assert.NotNil(t, e.GetRaw("@policy"))
	assert.NotNil(t, e.GetRaw("@mandate"))
	assert.NotNil(t, e.GetRaw("@parents"))

	assert.Equal(t, jp(m["@ctx:s"]), jp(e.GetRaw("@ctx")))
	assert.Equal(t, jp(m["@signature:o"]), jp(e.GetRaw("@signature")))
	assert.Equal(t, jp(m["@authority:o"]), jp(e.GetRaw("@authority")))
	assert.Equal(t, jp(m["@signer:o"]), jp(e.GetRaw("@signer")))
	assert.Equal(t, jp(m["@policy:o"]), jp(e.GetRaw("@policy")))
	assert.Equal(t, jp(m["@mandate:o"]), jp(e.GetRaw("@mandate")))
	assert.Equal(t, jp(m["@parents:a<s>"]), jp(e.GetRaw("@parents")))
}

// jp is a lazy approach to comparing the mess that is unmarshaling json when
// dealing with numbers
func jp(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ") // nolint
	return string(b)
}
