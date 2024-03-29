package blob_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/blob"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectmanagermock"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/resolvermock"
	"nimona.io/pkg/tilde"
)

func Test_requester_Request(t *testing.T) {
	pk, _ := crypto.NewEd25519PrivateKey()

	chunk1 := &blob.Chunk{
		Data: tilde.Data("ooh wee"),
	}
	chunk2 := &blob.Chunk{
		Data: tilde.Data("ooh lala"),
	}

	blob1 := &blob.Blob{
		Chunks: []tilde.Digest{
			object.MustMarshal(chunk1).Hash(),
			object.MustMarshal(chunk2).Hash(),
		},
	}

	peer1 := &peer.ConnectionInfo{
		Metadata: object.Metadata{
			Owner: pk.PublicKey().DID(),
		},
	}

	type fields struct {
		resolver func(*testing.T, *peer.ConnectionInfo) resolver.Resolver
		objmgr   func(*testing.T, *peer.ConnectionInfo) objectmanager.ObjectManager
	}
	type args struct {
		ctx  context.Context
		hash tilde.Digest
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       *blob.Blob
		wantChunks []*blob.Chunk
		wantErr    bool
	}{{
		name: "should pass",
		fields: fields{
			resolver: func(t *testing.T,
				pr *peer.ConnectionInfo,
			) resolver.Resolver {
				ctrl := gomock.NewController(t)
				mr := resolvermock.NewMockResolver(ctrl)
				mr.EXPECT().
					LookupByContent(gomock.Any(), gomock.Any()).
					Return([]*peer.ConnectionInfo{pr}, nil)
				return mr
			},
			objmgr: func(
				t *testing.T,
				pr *peer.ConnectionInfo,
			) objectmanager.ObjectManager {
				ctrl := gomock.NewController(t)
				mobm := objectmanagermock.NewMockObjectManager(ctrl)

				pubSub := objectmanager.NewObjectPubSub()
				pubSub.Publish(object.MustMarshal(blob1))

				mobm.EXPECT().Request(
					gomock.Any(),
					object.MustMarshal(blob1).Hash(),
					peer1.Metadata.Owner,
				).Return(object.MustMarshal(blob1), nil).MaxTimes(1)

				mobm.EXPECT().Request(
					gomock.Any(),
					object.MustMarshal(chunk1).Hash(),
					peer1.Metadata.Owner,
				).Return(object.MustMarshal(chunk1), nil)

				mobm.EXPECT().Request(
					gomock.Any(),
					object.MustMarshal(chunk2).Hash(),
					peer1.Metadata.Owner,
				).Return(object.MustMarshal(chunk2), nil)

				return mobm
			},
		},
		args: args{
			ctx:  context.Background(),
			hash: object.MustMarshal(blob1).Hash(),
		},
		want: blob1,
		wantChunks: []*blob.Chunk{
			chunk1,
			chunk2,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := blob.NewManager(
				tt.args.ctx,
				blob.WithObjectManager(tt.fields.objmgr(t, peer1)),
				blob.WithResolver(tt.fields.resolver(t, peer1)),
			)

			got, gotChunks, err := r.Request(tt.args.ctx, tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("requester.Request() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// we check the objects because they are easier to compare
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantChunks, gotChunks)
		})
	}
}

func Test_manager_ImportFromFile(t *testing.T) {
	chunk0 := &object.Object{
		Type: blob.ChunkType,
		Data: tilde.Map{
			"data": tilde.Data(
				"1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n13\n14" +
					"\n15\n16\n17\n18\n19\n20",
			),
		},
	}
	chunk1 := &object.Object{
		Type: blob.ChunkType,
		Data: tilde.Map{
			"data": tilde.Data(
				"\n21\n22\n23\n24\n25\n26\n27\n28\n29\n30\n31\n" +
					"32\n33\n34\n35\n36\n3",
			),
		},
	}
	chunk2 := &object.Object{
		Type: blob.ChunkType,
		Data: tilde.Map{
			"data": tilde.Data(
				"7\n38\n39\n40\n",
			),
		},
	}
	tests := []struct {
		name          string
		objectmanager func(*testing.T) objectmanager.ObjectManager
		path          string
		chunkSize     int
		want          *blob.Blob
		wantErr       bool
	}{{
		name:      "3 chunks, should pass",
		path:      "test-blob.bin",
		chunkSize: 50,
		objectmanager: func(t *testing.T) objectmanager.ObjectManager {
			c := gomock.NewController(t)
			m := objectmanagermock.NewMockObjectManager(c)
			m.EXPECT().
				Put(gomock.Any(), chunk0).
				MaxTimes(1)
			m.EXPECT().
				Put(gomock.Any(), chunk1).
				MaxTimes(1)
			m.EXPECT().
				Put(gomock.Any(), chunk2).
				MaxTimes(1)
			m.EXPECT().
				Put(gomock.Any(), &object.Object{
					Type: blob.BlobType,
					Data: tilde.Map{
						"chunks": tilde.DigestArray{
							chunk0.Hash(),
							chunk1.Hash(),
							chunk2.Hash(),
						},
					},
				}).
				MaxTimes(1)
			return m
		},
		want: &blob.Blob{
			Chunks: []tilde.Digest{
				chunk0.Hash(),
				chunk1.Hash(),
				chunk2.Hash(),
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := blob.NewManager(
				context.Background(),
				blob.WithChunkSize(tt.chunkSize),
				blob.WithImportWorkers(2),
				blob.WithObjectManager(tt.objectmanager(t)),
			)
			got, err := r.ImportFromFile(context.Background(), tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
