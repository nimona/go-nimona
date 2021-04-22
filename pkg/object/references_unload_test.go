package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
)

func TestUnloadReferences(t *testing.T) {
	f01 := &Object{
		Type: "f01",
		Data: Map{
			"f01": String("f01"),
		},
	}
	f02 := &Object{
		Type: "f02",
		Data: Map{
			"f02": String("f02"),
		},
	}
	f00 := &Object{
		Type: "f00",
		Data: Map{
			"f00": String("f00"),
			"f01": f01.CID(),
			"f02": f02.CID(),
		},
	}
	f00Full := &Object{
		Type: "f00",
		Data: Map{
			"f00": String("f00"),
			"f01": f01,
			"f02": f02,
		},
	}

	type args struct {
		ctx context.Context
		obj *Object
	}
	tests := []struct {
		name     string
		args     args
		want     *Object
		wantRefs []*Object
		wantErr  bool
	}{{
		name: "should pass, one object, no references",
		args: args{
			ctx: context.Background(),
			obj: f01,
		},
		want: f01,
	}, {
		name: "should pass, one object, two references",
		args: args{
			ctx: context.Background(),
			obj: f00Full,
		},
		want: f00,
		wantRefs: []*Object{
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
			require.EqualValues(t, tt.want, got)
		})
	}
}
