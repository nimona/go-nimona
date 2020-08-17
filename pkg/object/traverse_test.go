package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			"foo2:as/0": Int(1),
			"foo2:as/1": Int(2),
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
			"foo2:am/foo1:i": Int(1),
			"foo2:am/foo2:i": Int(2),
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
