package immutable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIterate(t *testing.T) {
	m := Map{}.
		Set("foo", "bar0").
		Set("foo", "bar1").
		Set("foo", "bar2").
		Set("not-foo", "not-bar0")

	p := map[string]interface{}{}
	m.Iterate(func(k string, v interface{}) {
		p[k] = v
	})

	assert.Equal(t, map[string]interface{}{
		"foo":     "bar2",
		"not-foo": "not-bar0",
	}, p)
}

func TestPlainMap(t *testing.T) {
	m := Map{}
	assert.Equal(t, nil, m.Value("foo"))
	iCalls := 0
	m.Iterate(func(_ string, _ interface{}) {
		iCalls++
	})
	assert.Equal(t, 0, iCalls)

	m = m.Set("foo", "bar")
	assert.Equal(t, "bar", m.Value("foo"))
	iCalls = 0
	m.Iterate(func(_ string, _ interface{}) {
		iCalls++
	})
	assert.Equal(t, 1, iCalls)

	nm := m.Set("foo", "nbar")
	assert.Equal(t, "bar", m.Value("foo"))
	assert.Equal(t, "nbar", nm.Value("foo"))
	iCalls = 0
	nm.Iterate(func(_ string, _ interface{}) {
		iCalls++
	})
	assert.Equal(t, 1, iCalls)

	nm = nm.Set("nfoo", "nbar")
	assert.Equal(t, "bar", m.Value("foo"))
	assert.Equal(t, "nbar", nm.Value("foo"))
	assert.Equal(t, "nbar", nm.Value("nfoo"))
	iCalls = 0
	nm.Iterate(func(_ string, _ interface{}) {
		iCalls++
	})
	assert.Equal(t, 2, iCalls)
}

func TestNewMap(t *testing.T) {
	m := NewMap()
	assert.Equal(t, nil, m.Value("foo"))
	iCalls := 0
	m.Iterate(func(_ string, _ interface{}) {
		iCalls++
	})
	assert.Equal(t, 0, iCalls)

	m = m.Set("foo", "bar")
	assert.Equal(t, "bar", m.Value("foo"))
	iCalls = 0
	m.Iterate(func(_ string, _ interface{}) {
		iCalls++
	})
	assert.Equal(t, 1, iCalls)

	nm := m.Set("foo", "nbar")
	assert.Equal(t, "bar", m.Value("foo"))
	assert.Equal(t, "nbar", nm.Value("foo"))
	iCalls = 0
	nm.Iterate(func(_ string, _ interface{}) {
		iCalls++
	})
	assert.Equal(t, 1, iCalls)

	nm = nm.Set("nfoo", "nbar")
	assert.Equal(t, "bar", m.Value("foo"))
	assert.Equal(t, "nbar", nm.Value("foo"))
	assert.Equal(t, "nbar", nm.Value("nfoo"))
	iCalls = 0
	nm.Iterate(func(_ string, _ interface{}) {
		iCalls++
	})
	assert.Equal(t, 2, iCalls)
}
