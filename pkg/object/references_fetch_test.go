package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
)

func TestFetchReferences(t *testing.T) {
	f00 := &Object{
		Data: map[string]interface{}{
			"f00:s": "f00",
			"f01:r": Hash("f01"),
			"f02:r": Hash("f02"),
		},
	}
	f01 := &Object{
		Data: map[string]interface{}{
			"f01:s": "f01",
		},
	}
	f02 := &Object{
		Data: map[string]interface{}{
			"f02:s": "f02",
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
		want    []*Object
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
		want: []*Object{
			f01,
		},
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
		want: []*Object{
			f00,
			f01,
			f02,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchWithReferences(
				tt.args.ctx,
				tt.args.getter,
				tt.args.objectHash,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				objs := []*Object{}
				for {
					obj, err := got.Read()
					if err != nil {
						break
					}
					objs = append(objs, obj)
				}
				require.Equal(t, len(tt.want), len(objs))
				assert.ElementsMatch(t, tt.want, objs)
			}
		})
	}
}
