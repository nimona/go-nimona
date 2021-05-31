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
			"f01": CID("f01"),
			"f02": CID("f02"),
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
		ctx       context.Context
		getter    GetterFunc
		objectCID CID
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
				cid CID,
			) (*Object, error) {
				if cid == "f01" {
					return f01, nil
				}
				return nil, errors.Error("not found")
			},
			objectCID: "f01",
		},
		want: f01,
	}, {
		name: "should pass, one object, two references",
		args: args{
			ctx: context.Background(),
			getter: func(
				ctx context.Context,
				cid CID,
			) (*Object, error) {
				switch cid {
				case "f00":
					return f00, nil
				case "f01":
					return f01, nil
				case "f02":
					return f02, nil
				}
				return nil, errors.Error("not found")
			},
			objectCID: "f00",
		},
		want: f00Full,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadReferences(
				tt.args.ctx,
				tt.args.objectCID,
				tt.args.getter,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
