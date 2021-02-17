package filesharing

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"nimona.io/pkg/blob"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/network"
	"nimona.io/pkg/networkmock"
	object "nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectmanagermock"
)

func Test_fileSharer_RequestTransfer(t *testing.T) {
	file1 := File{
		Name: "testfile",
	}

	type fields struct {
		objm func(
			*testing.T,
			context.Context,
		) objectmanager.ObjectManager
		net func(
			*testing.T,
			context.Context,
		) network.Network
	}
	type args struct {
		ctx     context.Context
		file    File
		peerKey crypto.PublicKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "should pass",
			args: args{
				ctx:  context.Background(),
				file: file1,
			},
			fields: fields{
				net: func(
					*testing.T,
					context.Context,
				) network.Network {
					return &networkmock.MockNetworkSimple{
						SendCalls: []error{
							nil,
						},
					}
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsh := New(
				nil,
				tt.fields.net(t, tt.args.ctx),
				"",
			)

			err := fsh.RequestTransfer(
				tt.args.ctx,
				&tt.args.file,
				tt.args.peerKey,
			)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func Test_fileSharer_Listen(t *testing.T) {
	file1 := File{
		Name: "testfile",
	}
	req := &TransferRequest{
		File: &file1,
	}

	type fields struct {
		objm func(
			*testing.T,
			context.Context,
		) objectmanager.ObjectManager
		net func(
			*testing.T,
			context.Context,
		) network.Network
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "receive one TransferRequest",
			fields: fields{
				objm: func(
					t *testing.T,
					ctx context.Context,
				) objectmanager.ObjectManager {
					ctrl := gomock.NewController(t)
					mobm := objectmanagermock.NewMockObjectManager(ctrl)

					mobm.EXPECT().Put(ctx, req.ToObject())
					return mobm
				},
				net: func(
					t *testing.T,
					ctx context.Context,
				) network.Network {
					m := &networkmock.MockNetworkSimple{
						SendCalls: []error{
							nil,
						},
						SubscribeCalls: []network.EnvelopeSubscription{
							&networkmock.MockSubscriptionSimple{
								Objects: []*network.Envelope{{
									Payload: req.ToObject(),
								}},
							},
						},
					}
					return m
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsh := New(
				tt.fields.objm(t, tt.args.ctx),
				tt.fields.net(t, tt.args.ctx),
				"",
			)

			events, err := fsh.Listen(tt.args.ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, events)

		})
	}
}

func Test_fileSharer_RequestFile(t *testing.T) {
	type fields struct {
		objm func(
			*testing.T,
			context.Context,
		) objectmanager.ObjectManager
		net func(
			*testing.T,
			context.Context,
		) network.Network
	}

	file1 := File{
		Name:   "testfile",
		Chunks: []object.Hash{"1234"},
	}
	req := &TransferRequest{
		File:  &file1,
		Nonce: "1234",
	}
	chunk1 := blob.Chunk{Data: []byte("asdf")}

	type args struct {
		ctx   context.Context
		hash  object.Hash
		nonce string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *os.File
		wantErr bool
	}{
		{
			name: "request one file",
			args: args{
				ctx:   context.TODO(),
				nonce: req.Nonce,
			},
			fields: fields{
				objm: func(
					t *testing.T,
					ctx context.Context,
				) objectmanager.ObjectManager {
					ctrl := gomock.NewController(t)
					mobm := objectmanagermock.NewMockObjectManager(ctrl)
					mobm.EXPECT().Request(ctx, file1.Chunks[0], gomock.Any()).
						Return(chunk1.ToObject(), nil)
					return mobm
				},
				net: func(
					t *testing.T,
					ctx context.Context,
				) network.Network {
					m := &networkmock.MockNetworkSimple{
						SendCalls: []error{
							nil,
						},
						SubscribeCalls: []network.EnvelopeSubscription{
							&networkmock.MockSubscriptionSimple{
								Objects: []*network.Envelope{{
									Payload: req.ToObject(),
								}},
							},
						},
					}
					return m
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsh := New(
				tt.fields.objm(t, tt.args.ctx),
				tt.fields.net(t, tt.args.ctx),
				"",
			)

			f := fsh.(*fileSharer)
			f.incomingTransfer[req.Nonce] = &Transfer{
				Request: *req,
			}

			events, err := fsh.Listen(tt.args.ctx)
			assert.NoError(t, err)
			assert.NotNil(t, events)

			file, err := fsh.RequestFile(tt.args.ctx, tt.args.hash, tt.args.nonce)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, file)
		})
	}
}
