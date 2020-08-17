package object

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_resolvePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		value   Value
		want    Value
		wantErr bool
	}{{
		name: "should fail, invalid path",
		value: Map{}.
			Set("foo:s", String("bar")),
		path:    "",
		wantErr: true,
	}, {
		name:    "should fail, invalid value",
		value:   Int(5),
		path:    "foo:s",
		wantErr: true,
	}, {
		name: "should pass, simple map",
		value: Map{}.
			Set("foo:s", String("bar")),
		path: "foo:s",
		want: String("bar"),
	}, {
		name: "should fail, simple map",
		value: Map{}.
			Set("foo:s", String("bar")),
		path:    "foo-x:s",
		wantErr: true,
	}, {
		name: "should pass, nested map",
		value: Map{}.
			Set("foo:s", String("bar")).
			Set("nested:m",
				Map{}.
					Set("foo:s", String("nested-bar")),
			),
		path: "nested:m/foo:s",
		want: String("nested-bar"),
	}, {
		name: "should fail, map, list -1, map",
		value: Map{}.
			Set("foo:s", String("bar")).
			Set("list:am", List{}.
				Append(Map{}),
			),
		path:    "list:am/-1/nested:m/foo:s",
		wantErr: true,
	}, {
		name: "should fail, map, list 3, map",
		value: Map{}.
			Set("foo:s", String("bar")).
			Set("list:am", List{}.
				Append(Map{}),
			),
		path:    "list:am/3/nested:m/foo:s",
		wantErr: true,
	}, {
		name: "should pass, map, list 0, map",
		value: Map{}.
			Set("foo:s", String("bar")).
			Set("list:am", List{}.
				Append(
					Map{}.
						Set("nested:m",
							Map{}.
								Set("foo:s", String("nested-bar0")),
						),
				).
				Append(
					Map{}.
						Set("nested:m",
							Map{}.
								Set("foo:s", String("nested-bar1")),
						),
				),
			),
		path: "list:am/0/nested:m/foo:s",
		want: String("nested-bar0"),
	}, {
		name: "should pass, map, list 1, map",
		value: Map{}.
			Set("foo:s", String("bar")).
			Set("list:am", List{}.
				Append(
					Map{}.
						Set("nested:m",
							Map{}.
								Set("foo:s", String("nested-bar0")),
						),
				).
				Append(
					Map{}.
						Set("nested:m",
							Map{}.
								Set("foo:s", String("nested-bar1")),
						),
				),
			),
		path: "list:am/1/nested:m/foo:s",
		want: String("nested-bar1"),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePath(tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolvePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("resolvePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setPath(t *testing.T) {
	tests := []struct {
		name    string
		target  Value
		path    string
		value   Value
		want    Value
		wantErr bool
	}{{
		name: "should pass, set map",
		target: Map{}.
			Set("foo:s", String("bar")),
		path:  "foo:s",
		value: String("BAR"),
		want: Map{}.
			Set("foo:s", String("BAR")),
	}, {
		name: "should pass, set list",
		target: List{}.
			Append(String("bar")),
		path:  "0",
		value: String("BAR"),
		want: List{}.
			Append(String("BAR")),
	}, {
		name: "should pass, set nested list, map",
		target: Map{}.
			Set("foo:s", String("bar")).
			Set("list:am", List{}.
				Append(
					Map{}.
						Set("nested:m",
							Map{}.
								Set("foo:s", String("nested-bar0")),
						),
				),
			),
		path:  "list:am/0/nested:m/foo:s",
		value: String("BAR"),
		want: Map{}.
			Set("foo:s", String("bar")).
			Set("list:am", List{}.
				Append(
					Map{}.
						Set("nested:m",
							Map{}.
								Set("foo:s", String("BAR")),
						),
				),
			),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := setPath(tt.target, tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("setPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want.PrimitiveHinted(), got.PrimitiveHinted())
		})
	}
}
