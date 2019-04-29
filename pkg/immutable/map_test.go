package immutable

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrimitive(t *testing.T) {
	m := Map{}.
		Set("foo", Value{stringValue{"bar0"}}).
		Set("foo", Value{stringValue{"bar1"}}).
		Set("foo", Value{stringValue{"bar2"}}).
		Set("not-foo", Value{stringValue{"not-bar0"}}).
		Set("nested-map", Value{mapValue{
			Map{}.Set("nested-foo", Value{stringValue{"nested-bar"}}),
		}})

	p := m.Primitive()
	assert.Equal(t, map[string]interface{}{
		"foo":     "bar2",
		"not-foo": "not-bar0",
		"nested-map": map[string]interface{}{
			"nested-foo": "nested-bar",
		},
	}, p)

	h := m.PrimitiveHinted()
	assert.Equal(t, map[string]interface{}{
		"foo:s":     "bar2",
		"not-foo:s": "not-bar0",
		"nested-map:o": map[string]interface{}{
			"nested-foo:s": "nested-bar",
		},
	}, h)
}

func TestMap(t *testing.T) {
	m := Map{}
	assert.Equal(t, "", m.Value("foo").StringValue())
	iCalls := 0
	m.Iterate(func(_ string, _ Value) {
		iCalls++
	})
	assert.Equal(t, 0, iCalls)

	m = m.Set("foo", Value{stringValue{"bar"}})
	assert.Equal(t, "bar", m.Value("foo").StringValue())
	iCalls = 0
	m.Iterate(func(_ string, _ Value) {
		iCalls++
	})
	assert.Equal(t, 1, iCalls)

	nm := m.Set("foo", Value{stringValue{"nbar"}})
	assert.Equal(t, "bar", m.Value("foo").StringValue())
	assert.Equal(t, "nbar", nm.Value("foo").StringValue())
	iCalls = 0
	nm.Iterate(func(_ string, _ Value) {
		iCalls++
	})
	assert.Equal(t, 1, iCalls)

	nm = nm.Set("nfoo", Value{stringValue{"nbar"}})
	assert.Equal(t, "bar", m.Value("foo").StringValue())
	assert.Equal(t, "nbar", nm.Value("foo").StringValue())
	assert.Equal(t, "nbar", nm.Value("nfoo").StringValue())
	iCalls = 0
	nm.Iterate(func(_ string, _ Value) {
		iCalls++
	})
	assert.Equal(t, 2, iCalls)
}
