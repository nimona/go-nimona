package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
)

func TestUnloadReferences(t *testing.T) {
	f00 := Object{}.
		Set("f00:s", "f00").
		Set("f01:r", "oh1.CpcViyHidQoytZo8d6jmncFzWYsXSG5nHBFZhveEjH6r").
		Set("f02:r", "oh1.ABm8HB1oAZGq5TdvoGo416s71FwoZJdw3jk5zU4QRbiK")
	f01 := Object{}.
		SetType("f01").
		Set("f01:s", "f01")
	f02 := Object{}.
		SetType("f02").
		Set("f02:s", "f02")

	f00Full := Object{}.
		Set("f00:s", "f00").
		Set("f01:m", f01.Raw()).
		Set("f02:m", f02.Raw())

	type args struct {
		ctx context.Context
		obj Object
	}
	tests := []struct {
		name     string
		args     args
		want     *Object
		wantRefs []Object
		wantErr  bool
	}{{
		name: "should pass, one object, no references",
		args: args{
			ctx: context.Background(),
			obj: f01,
		},
		want: &f01,
	}, {
		name: "should pass, one object, two references",
		args: args{
			ctx: context.Background(),
			obj: f00Full,
		},
		want: &f00,
		wantRefs: []Object{
			f01,
			f02,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotRefs, err := UnloadReferences(
				tt.args.ctx,
				tt.args.obj,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want.ToMap(), got.ToMap())
			require.Equal(t, len(tt.wantRefs), len(gotRefs))
			for i := 0; i < len(tt.wantRefs); i++ {
				assert.Equal(
					t,
					tt.wantRefs[i].ToMap(),
					gotRefs[i].ToMap(),
					"for index %d", i,
				)
			}
		})
	}
}
