package immutable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromPrimitive(t *testing.T) {
	em := map[string]interface{}{
		"foo:s":     "bar2",
		"not-foo:s": "not-bar0",
		"nested-map:o": map[string]interface{}{
			"nested-foo:s": "nested-bar",
		},
		"foos:as": []string{
			"foo0",
			"foo1",
			"foo2",
		},
	}

	v := AnyToValue(":o", em)
	m := v.PrimitiveHinted()
	require.Equal(t, em, m)
}

func TestMapPrimitive(t *testing.T) {
	l := List{
		hint: "as",
	}
	l = l.Append(String("foo0"))
	l = l.Append(String("foo1"))
	l = l.Append(String("foo2"))
	m := Map{}.
		Set("foo", String("bar0")).
		Set("foo", String("bar1")).
		Set("foo", String("bar2")).
		Set("not-foo", String("not-bar0")).
		Set("nested-map", Map{}.Set("nested-foo", String("nested-bar"))).
		Set("foos", l)

	h := m.PrimitiveHinted()
	assert.Equal(t, map[string]interface{}{
		"foo:s":     "bar2",
		"not-foo:s": "not-bar0",
		"nested-map:o": map[string]interface{}{
			"nested-foo:s": "nested-bar",
		},
		"foos:as": []string{
			"foo0",
			"foo1",
			"foo2",
		},
	}, h)
}

func TestMap(t *testing.T) {
	m := Map{}
	assert.Equal(t, nil, m.Value("foo"))
	iCalls := 0
	m.Iterate(func(_ string, _ Value) {
		iCalls++
	})
	assert.Equal(t, 0, iCalls)

	m = m.Set("foo", String("bar"))
	assert.Equal(t, "bar", m.Value("foo").PrimitiveHinted().(string))
	iCalls = 0
	m.Iterate(func(_ string, _ Value) {
		iCalls++
	})
	assert.Equal(t, 1, iCalls)

	nm := m.Set("foo", String("nbar"))
	assert.Equal(t, "bar", m.Value("foo").PrimitiveHinted().(string))
	assert.Equal(t, "nbar", nm.Value("foo").PrimitiveHinted().(string))
	iCalls = 0
	nm.Iterate(func(_ string, _ Value) {
		iCalls++
	})
	assert.Equal(t, 1, iCalls)

	nm = nm.Set("nfoo", String("nbar"))
	assert.Equal(t, "bar", m.Value("foo").PrimitiveHinted().(string))
	assert.Equal(t, "nbar", nm.Value("foo").PrimitiveHinted().(string))
	assert.Equal(t, "nbar", nm.Value("nfoo").PrimitiveHinted().(string))
	iCalls = 0
	nm.Iterate(func(_ string, _ Value) {
		iCalls++
	})
	assert.Equal(t, 2, iCalls)
}
