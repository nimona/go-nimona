package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Traverse(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{{
		name: "should pass, simple value",
		args: args{
			1,
		},
		want: map[string]interface{}{
			"": 1,
		},
	}, {
		name: "should pass, simple list",
		args: args{
			v: []int64{0, 1, 2},
		},
		want: map[string]interface{}{
			"":  []int64{0, 1, 2},
			"0": int64(0),
			"1": int64(1),
			"2": int64(2),
		},
	}, {
		name: "should pass, simple map",
		args: args{
			v: map[string]interface{}{
				"foo0:s": "bar0",
				"foo1:s": "bar1",
			},
		},
		want: map[string]interface{}{
			"": map[string]interface{}{
				"foo0:s": "bar0",
				"foo1:s": "bar1",
			},
			"foo0:s": "bar0",
			"foo1:s": "bar1",
		},
	}, {
		name: "should pass, complex map",
		args: args{
			v: map[string]interface{}{
				"foo0:s":  "bar0",
				"foo1:s":  "bar1",
				"foo2:as": []int64{1, 2, 3},
				"foo3:i":  4,
			},
		},
		want: map[string]interface{}{
			"": map[string]interface{}{
				"foo0:s":  "bar0",
				"foo1:s":  "bar1",
				"foo2:as": []int64{1, 2, 3},
				"foo3:i":  4,
			},
			"foo0:s":    "bar0",
			"foo1:s":    "bar1",
			"foo2:as":   []int64{1, 2, 3},
			"foo2:as/0": int64(1),
			"foo2:as/1": int64(2),
			"foo2:as/2": int64(3),
			"foo3:i":    4,
		},
	}, {
		name: "should pass, nested map",
		args: args{
			v: map[string]interface{}{
				"foo0:s": "bar0",
				"foo1:s": "bar1",
				"foo2:am": map[string]interface{}{
					"foo1:i": int64(1),
					"foo2:i": int64(2),
					"foo3:i": int64(3),
				},
				"foo3:i": int64(4),
			},
		},
		want: map[string]interface{}{
			"": map[string]interface{}{
				"foo0:s": "bar0",
				"foo1:s": "bar1",
				"foo2:am": map[string]interface{}{
					"foo1:i": int64(1),
					"foo2:i": int64(2),
					"foo3:i": int64(3),
				},
				"foo3:i": int64(4),
			},
			"foo0:s": "bar0",
			"foo1:s": "bar1",
			"foo2:am": map[string]interface{}{
				"foo1:i": int64(1),
				"foo2:i": int64(2),
				"foo3:i": int64(3),
			},
			"foo2:am/foo1:i": int64(1),
			"foo2:am/foo2:i": int64(2),
			"foo2:am/foo3:i": int64(3),
			"foo3:i":         int64(4),
		},
	}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			res := map[string]interface{}{}
			Traverse(tt.args.v, func(k string, v interface{}) bool {
				res[k] = v
				return true
			})
			assert.Equal(t, tt.want, res)
		})
	}
}
