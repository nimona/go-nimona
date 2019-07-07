package immutable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromPrimitive(t *testing.T) {
	em := map[interface{}]interface{}{
		"foo:s":     "bar2",
		"not-foo:s": "not-bar0",
		"nested-map:o": map[interface{}]interface{}{
			"nested-foo:s": "nested-bar",
		},
		"foos:as": []interface{}{
			"foo0",
			"foo1",
			"foo2",
		},
	}

	v := AnyToValue(":o", em)

	// spew.Dump(v)
	m := v.PrimitiveHinted()

	require.Equal(t, em, m)
}

func TestMapPrimitive(t *testing.T) {
	l := List{}
	l = l.Append(Value{stringValue{"foo0"}})
	l = l.Append(Value{stringValue{"foo1"}})
	l = l.Append(Value{stringValue{"foo2"}})
	m := Map{}.
		Set("foo", Value{stringValue{"bar0"}}).
		Set("foo", Value{stringValue{"bar1"}}).
		Set("foo", Value{stringValue{"bar2"}}).
		Set("not-foo", Value{stringValue{"not-bar0"}}).
		Set("nested-map", Value{mapValue{
			Map{}.Set("nested-foo", Value{stringValue{"nested-bar"}}),
		}}).
		Set("foos", Value{listValue{"as", l}})

	p := m.primitive()
	assert.Equal(t, map[interface{}]interface{}{
		"foo":     "bar2",
		"not-foo": "not-bar0",
		"nested-map": map[interface{}]interface{}{
			"nested-foo": "nested-bar",
		},
		"foos": []interface{}{
			"foo0",
			"foo1",
			"foo2",
		},
	}, p)

	h := m.primitiveHinted()
	assert.Equal(t, map[interface{}]interface{}{
		"foo:s":     "bar2",
		"not-foo:s": "not-bar0",
		"nested-map:o": map[interface{}]interface{}{
			"nested-foo:s": "nested-bar",
		},
		"foos:as": []interface{}{
			"foo0",
			"foo1",
			"foo2",
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
