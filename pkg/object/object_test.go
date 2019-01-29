package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectMethods(t *testing.T) {
	m := map[string]interface{}{
		"@ctx": "ctx-value",
		"@signature": FromMap(map[string]interface{}{
			"@ctx": "-signature",
		}),
		"@authority": FromMap(map[string]interface{}{
			"@ctx": "-authority",
		}),
		"@signer": FromMap(map[string]interface{}{
			"@ctx": "-signer",
		}),
		"@policy": FromMap(map[string]interface{}{
			"@ctx": "-policy",
		}),
		"@mandate": FromMap(map[string]interface{}{
			"@ctx": "-mandate",
		}),
		"@parents": []string{"parents-value"},
	}

	o := FromMap(m)

	assert.Equal(t, m["@ctx"], o.GetRaw("@ctx"))
	assert.Equal(t, m["@signature"], o.GetRaw("@signature"))
	assert.Equal(t, m["@authority"], o.GetRaw("@authority"))
	assert.Equal(t, m["@signer"], o.GetRaw("@signer"))
	assert.Equal(t, m["@policy"], o.GetRaw("@policy"))
	assert.Equal(t, m["@mandate"], o.GetRaw("@mandate"))
	assert.Equal(t, m["@parents"], o.GetRaw("@parents"))

	n := &Object{}

	n.SetRaw("@ctx", m["@ctx"])
	n.SetRaw("@signature", m["@signature"])
	n.SetRaw("@authority", m["@authority"])
	n.SetRaw("@signer", m["@signer"])
	n.SetRaw("@policy", m["@policy"])
	n.SetRaw("@mandate", m["@mandate"])
	n.SetRaw("@parents", m["@parents"])

	assert.Equal(t, m["@ctx"], n.GetRaw("@ctx"))
	assert.Equal(t, m["@signature"], n.GetRaw("@signature"))
	assert.Equal(t, m["@authority"], n.GetRaw("@authority"))
	assert.Equal(t, m["@signer"], n.GetRaw("@signer"))
	assert.Equal(t, m["@policy"], n.GetRaw("@policy"))
	assert.Equal(t, m["@mandate"], n.GetRaw("@mandate"))
	assert.Equal(t, m["@parents"], n.GetRaw("@parents"))

	e := &Object{}

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

	assert.Equal(t, m["@ctx"], e.GetRaw("@ctx"))
	assert.Equal(t, m["@signature"], e.GetRaw("@signature"))
	assert.Equal(t, m["@authority"], e.GetRaw("@authority"))
	assert.Equal(t, m["@signer"], e.GetRaw("@signer"))
	assert.Equal(t, m["@policy"], e.GetRaw("@policy"))
	assert.Equal(t, m["@mandate"], e.GetRaw("@mandate"))
	assert.Equal(t, m["@parents"], e.GetRaw("@parents"))
}
