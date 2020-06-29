package objectloader

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/errors"
	"nimona.io/pkg/exchange"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectstore"
	"nimona.io/pkg/objectstore/objectstoremock"
	"nimona.io/pkg/resolver"
)

var (
	testObject0 = object.Object{}.
			SetType("bar").
			Set("foo:s", object.String("bar"))
	testObject0Ref = object.Ref("oh1.D5Uoh8k7vHe3xzcnbw5Z5XBsetqMmmBuijU2ampLt4do")
	testObject1    = object.Object{}.
			SetType("foo").
			Set("foo:s", object.String("bar")).
			Set("obj:m", testObject0.ToMap())
	testObject1Unloaded = object.Object{}.
				SetType("foo").
				Set("foo:s", object.String("bar")).
				Set("obj:r", testObject0Ref)
)

func Test_loader_Load(t *testing.T) {
	type fields struct {
		store func(*testing.T) objectstore.Store
	}
	type args struct {
		obj  object.Object
		opts []option
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *object.Object
		wantErr bool
	}{{
		name: "should pass, 1 object loaded",
		args: args{
			obj: testObject1Unloaded,
		},
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
				m.EXPECT().Get(
					object.Hash(testObject0Ref),
				).Return(
					testObject0,
					nil,
				)
				return m
			},
		},
		want: &testObject1,
	}, {
		name: "should fail, 1 object failed to load from store",
		args: args{
			obj: testObject1Unloaded,
		},
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				m := objectstoremock.NewMockStore(
					gomock.NewController(t),
				)
				m.EXPECT().Get(
					object.Hash(testObject0Ref),
				).Return(
					object.Object{},
					errors.New("something bad"),
				)
				return m
			},
		},
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &loader{
				store: tt.fields.store(t),
			}
			got, err := l.Load(tt.args.obj, tt.args.opts...)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if tt.want != nil {
				assert.Equal(t, tt.want.ToMap(), got.ToMap())
			}
		})
	}
}

func Test_loader_Unload(t *testing.T) {
	type fields struct {
		store    func(*testing.T) objectstore.Store
		resolver func(*testing.T) resolver.Resolver
		exchange func(*testing.T) exchange.Exchange
	}
	type args struct {
		obj  object.Object
		opts []option
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		want              *object.Object
		wantUnloadedCount int
		wantErr           bool
	}{{
		name: "should pass, 1 object unloaded",
		args: args{
			obj: testObject1,
		},
		fields: fields{
			store: func(t *testing.T) objectstore.Store {
				return nil
			},
			resolver: func(t *testing.T) resolver.Resolver {
				return nil
			},
			exchange: func(t *testing.T) exchange.Exchange {
				return nil
			},
		},
		want:              &testObject1Unloaded,
		wantUnloadedCount: 1,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &loader{
				store: tt.fields.store(t),
			}
			got, gotUnloaded, err := l.Unload(tt.args.obj, tt.args.opts...)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want.ToMap(), got.ToMap())
			assert.Len(t, gotUnloaded, tt.wantUnloadedCount)
		})
	}
}
