package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromPrimitive(t *testing.T) {
	em := map[string]interface{}{
		"foo:s":     "bar2",
		"not-foo:s": "not-bar0",
		"nested-map:m": map[string]interface{}{
			"nested-foo:s": "nested-bar",
		},
		"foos:as": []string{
			"foo0",
			"foo1",
			"foo2",
		},
	}

	v := AnyToValue(":m", em)
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
		Set("foo:s", String("bar0")).
		Set("foo:s", String("bar1")).
		Set("foo:s", String("bar2")).
		Set("not-foo:s", String("not-bar0")).
		Set("nested-map:m", Map{}.Set("nested-foo:s", String("nested-bar"))).
		Set("foos:as", l)

	h := m.PrimitiveHinted()
	assert.Equal(t, map[string]interface{}{
		"foo:s":     "bar2",
		"not-foo:s": "not-bar0",
		"nested-map:m": map[string]interface{}{
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
	assert.Equal(t, nil, m.Value("foo:s"))
	iCalls := 0
	m.Iterate(func(_ string, _ Value) bool {
		iCalls++
		return true
	})
	assert.Equal(t, 0, iCalls)

	m = m.Set("foo:s", String("bar"))
	assert.Equal(t, "bar", m.Value("foo:s").PrimitiveHinted().(string))
	iCalls = 0
	m.Iterate(func(_ string, _ Value) bool {
		iCalls++
		return true
	})
	assert.Equal(t, 1, iCalls)

	nm := m.Set("foo:s", String("nbar"))
	assert.Equal(t, "bar", m.Value("foo:s").PrimitiveHinted().(string))
	assert.Equal(t, "nbar", nm.Value("foo:s").PrimitiveHinted().(string))
	iCalls = 0
	nm.Iterate(func(_ string, _ Value) bool {
		iCalls++
		return true
	})
	assert.Equal(t, 1, iCalls)

	nm = nm.Set("nfoo:s", String("nbar"))
	assert.Equal(t, "bar", m.Value("foo:s").PrimitiveHinted().(string))
	assert.Equal(t, "nbar", nm.Value("foo:s").PrimitiveHinted().(string))
	assert.Equal(t, "nbar", nm.Value("nfoo:s").PrimitiveHinted().(string))
	iCalls = 0
	nm.Iterate(func(_ string, _ Value) bool {
		iCalls++
		return true
	})
	assert.Equal(t, 2, iCalls)
}

func Test_Traverse(t *testing.T) {
	type args struct {
		v Value
	}
	tests := []struct {
		name      string
		args      args
		stopAfter int
		want      map[string]Value
	}{{
		name: "should pass, simple value",
		args: args{
			Int(1),
		},
		want: map[string]Value{
			"": Int(1),
		},
	}, {
		name: "should pass, simple list",
		args: args{
			v: List{}.
				Append(Int(0)).
				Append(Int(1)).
				Append(Int(2)),
		},
		want: map[string]Value{
			"": List{}.
				Append(Int(0)).
				Append(Int(1)).
				Append(Int(2)),
			"0": Int(0),
			"1": Int(1),
			"2": Int(2),
		},
	}, {
		name: "should pass, simple map",
		args: args{
			v: Map{}.
				Set("foo0:s", String("bar0")).
				Set("foo1:s", String("bar1")),
		},
		want: map[string]Value{
			"": Map{}.
				Set("foo0:s", String("bar0")).
				Set("foo1:s", String("bar1")),
			"foo0:s": String("bar0"),
			"foo1:s": String("bar1"),
		},
	}, {
		name: "should pass, simple map, stop after 2",
		args: args{
			v: Map{}.
				Set("foo0:s", String("bar0")).
				Set("foo1:s", String("bar1")),
		},
		stopAfter: 2,
		want: map[string]Value{
			"": Map{}.
				Set("foo0:s", String("bar0")).
				Set("foo1:s", String("bar1")),
			"foo0:s": String("bar0"),
		},
	}, {
		name: "should pass, complex map, stop after 6",
		args: args{
			v: Map{}.
				Set("foo0:s", String("bar0")).
				Set("foo1:s", String("bar1")).
				Set("foo2:as", List{}.
					Append(Int(1)).
					Append(Int(2)).
					Append(Int(3)),
				).
				Set("foo3:i", Int(4)),
		},
		stopAfter: 6,
		want: map[string]Value{
			"": Map{}.
				Set("foo0:s", String("bar0")).
				Set("foo1:s", String("bar1")).
				Set("foo2:as", List{}.
					Append(Int(1)).
					Append(Int(2)).
					Append(Int(3)),
				).
				Set("foo3:i", Int(4)),
			"foo0:s": String("bar0"),
			"foo1:s": String("bar1"),
			"foo2:as": List{}.
				Append(Int(1)).
				Append(Int(2)).
				Append(Int(3)),
			"foo2:as.0": Int(1),
			"foo2:as.1": Int(2),
		},
	}, {
		name: "should pass, nested map, stop after 6",
		args: args{
			v: Map{}.
				Set("foo0:s", String("bar0")).
				Set("foo1:s", String("bar1")).
				Set("foo2:am", Map{}.
					Set("foo1:i", Int(1)).
					Set("foo2:i", Int(2)).
					Set("foo3:i", Int(3)),
				).
				Set("foo3:i", Int(4)),
		},
		stopAfter: 6,
		want: map[string]Value{
			"": Map{}.
				Set("foo0:s", String("bar0")).
				Set("foo1:s", String("bar1")).
				Set("foo2:am", Map{}.
					Set("foo1:i", Int(1)).
					Set("foo2:i", Int(2)).
					Set("foo3:i", Int(3)),
				).
				Set("foo3:i", Int(4)),
			"foo0:s": String("bar0"),
			"foo1:s": String("bar1"),
			"foo2:am": Map{}.
				Set("foo1:i", Int(1)).
				Set("foo2:i", Int(2)).
				Set("foo3:i", Int(3)),
			"foo2:am.foo1:i": Int(1),
			"foo2:am.foo2:i": Int(2),
		},
	}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			res := map[string]Value{}
			Traverse(tt.args.v, func(k string, v Value) bool {
				res[k] = v
				if tt.stopAfter > 0 && len(res) == tt.stopAfter {
					return false
				}
				return true
			})
			assert.Equal(t, tt.want, res)
		})
	}
}
