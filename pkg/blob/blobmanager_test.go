package blob_test

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/blob"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/localpeer"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectmanagermock"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/resolvermock"
	"nimona.io/pkg/sqlobjectstore"
)

func Test_requester_Request(t *testing.T) {
	localPeer1 := newPeer()

	blob1 := &blob.Blob{}
	chunk1 := &blob.Chunk{Data: []byte("ooh wee")}
	chunk2 := &blob.Chunk{Data: []byte("ooh lala")}

	blob1.Chunks = []*blob.Chunk{chunk1, chunk2}

	peer1 := &peer.Peer{
		Metadata: object.Metadata{
			Owner: localPeer1.GetPrimaryIdentityKey().PublicKey(),
			// Addresses: n2.Addresses(),
		},
	}

	type fields struct {
		store    *sqlobjectstore.Store
		resolver func(*testing.T, *peer.Peer) resolver.Resolver
		objmgr   func(*testing.T, *peer.Peer) objectmanager.ObjectManager
	}
	type args struct {
		ctx  context.Context
		hash object.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *blob.Blob
		wantErr bool
	}{
		{
			name: "should pass",
			fields: fields{
				resolver: func(t *testing.T, pr *peer.Peer) resolver.Resolver {
					ctrl := gomock.NewController(t)
					mr := resolvermock.NewMockResolver(ctrl)
					mr.EXPECT().Lookup(gomock.Any(), gomock.Any()).DoAndReturn(
						func(
							ctx context.Context,
							opts ...resolver.LookupOption,
						) (<-chan *peer.Peer, error) {
							ch := make(chan *peer.Peer)
							go func() {
								ch <- pr
							}()
							return ch, nil
						})
					return mr
				},
				objmgr: func(t *testing.T, pr *peer.Peer) objectmanager.ObjectManager {
					ctrl := gomock.NewController(t)
					mobm := objectmanagermock.NewMockObjectManager(ctrl)

					pubSub := objectmanager.NewObjectPubSub()
					pubSub.Publish(blob1.ToObject())

					obj, _, err := object.UnloadReferences(context.TODO(), blob1.ToObject())

					assert.Len(t, blob1.Chunks, 2)
					assert.NoError(t, err)

					mobm.EXPECT().Request(
						gomock.Any(),
						blob1.ToObject().Hash(),
						peer1,
						true,
					).Return(obj, nil).MaxTimes(1)

					for _, ch := range blob1.Chunks {
						o := ch.ToObject()
						mobm.EXPECT().Request(
							gomock.Any(),
							ch.ToObject().Hash(),
							peer1,
							true,
						).Return(&o, nil)
					}

					mobm.EXPECT().Subscribe(
						gomock.Any(),
					).Return(
						pubSub.Subscribe(),
					)

					return mobm
				},
			},
			args: args{
				ctx:  context.Background(),
				hash: blob1.ToObject().Hash(),
			},
			want: blob1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := blob.NewRequester(
				tt.args.ctx,
				blob.WithObjectManager(tt.fields.objmgr(t, peer1)),
				blob.WithResolver(tt.fields.resolver(t, peer1)),
				blob.WithStore(tt.fields.store),
			)

			got, err := r.Request(tt.args.ctx, tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("requester.Request() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// we check the objects because they are easier to compare
			if !reflect.DeepEqual(
				got.ToObject().ToMap(),
				tt.want.ToObject().ToMap(),
			) {
				t.Errorf("requester.Request() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newPeer() localpeer.LocalPeer {
	pk, _ := crypto.GenerateEd25519PrivateKey()
	pk1, _ := crypto.GenerateEd25519PrivateKey()

	kc := localpeer.New()
	kc.PutPrimaryPeerKey(pk)
	kc.PutPrimaryIdentityKey(pk1)

	return kc
}

func TestUnload(t *testing.T) {
	blob1 := &blob.Blob{}
	chunk1 := &blob.Chunk{Data: []byte("ooh wee")}
	chunk2 := &blob.Chunk{Data: []byte("ooh lala")}

	blob1.Chunks = []*blob.Chunk{chunk1, chunk2}

	obj, _, err := object.UnloadReferences(context.TODO(), blob1.ToObject())
	assert.NoError(t, err)
	assert.NotNil(t, obj)

	refs := object.GetReferences(*obj)

	assert.Contains(t, refs, chunk1.ToObject().Hash())
	assert.Contains(t, refs, chunk2.ToObject().Hash())
}