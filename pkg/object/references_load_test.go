package object

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
)

func TestLoadReferences(t *testing.T) {
	f00 := &Object{
		Type: "f00",
		Data: Map{
			"f00": String("f00"),
			"f01": Hash("f01"),
			"f02": Hash("f02"),
		},
	}
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
	f00Full := &Object{
		Type: "f00",
		Data: Map{
			"f00": String("f00"),
			"f01": f01,
			"f02": f02,
		},
	}

	type args struct {
		ctx        context.Context
		getter     GetterFunc
		objectHash Hash
	}
	tests := []struct {
		name    string
		args    args
		want    *Object
		wantErr bool
	}{{
		name: "should pass, one object, no references",
		args: args{
			ctx: context.Background(),
			getter: func(
				ctx context.Context,
				hash Hash,
			) (*Object, error) {
				if hash == "f01" {
					return f01, nil
				}
				return nil, errors.New("not found")
			},
			objectHash: "f01",
		},
		want: f01,
	}, {
		name: "should pass, one object, two references",
		args: args{
			ctx: context.Background(),
			getter: func(
				ctx context.Context,
				hash Hash,
			) (*Object, error) {
				switch hash {
				case "f00":
					return f00, nil
				case "f01":
					return f01, nil
				case "f02":
					return f02, nil
				}
				return nil, errors.New("not found")
			},
			objectHash: "f00",
		},
		want: f00Full,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadReferences(
				tt.args.ctx,
				tt.args.objectHash,
				tt.args.getter,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want.ToMap(), got.ToMap())
		})
	}
}
