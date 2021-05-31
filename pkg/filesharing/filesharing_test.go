package filesharing_test

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"nimona.io/pkg/blob"
	"nimona.io/pkg/filesharing"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/network"
	"nimona.io/pkg/networkmock"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/objectmanagermock"
)

func Test_fileSharer_RequestTransfer(t *testing.T) {
	file1 := filesharing.File{
		Name: "testfile",
	}

	type fields struct {
		net func(
			*testing.T,
			context.Context,
		) network.Network
	}
	type args struct {
		ctx     context.Context
		file    filesharing.File
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
			fsh := filesharing.New(
				nil,
				tt.fields.net(t, tt.args.ctx),
				"",
			)

			nonce, err := fsh.RequestTransfer(
				tt.args.ctx,
				&tt.args.file,
				tt.args.peerKey,
			)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NotEmpty(t, nonce)
			assert.NoError(t, err)
		})
	}
}

func Test_fileSharer_Listen(t *testing.T) {
	file1 := filesharing.File{
		Name: "testfile",
	}
	req := &filesharing.TransferRequest{
		File: file1,
	}

	type fields struct {
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
									Payload: object.MustMarshal(req),
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
			fsh := filesharing.New(
				nil,
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

	file1 := filesharing.File{
		Name:   "testfile",
		Chunks: []object.CID{"1234"},
	}
	req := &filesharing.TransferRequest{
		File:  file1,
		Nonce: "1234",
	}
	chunk1 := blob.Chunk{Data: []byte("asdf")}

	type args struct {
		ctx   context.Context
		CID   object.CID
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
						Return(object.MustMarshal(chunk1), nil)
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
									Payload: object.MustMarshal(req),
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
			fsh := filesharing.New(
				tt.fields.objm(t, tt.args.ctx),
				tt.fields.net(t, tt.args.ctx),
				"",
			)

			events, err := fsh.Listen(tt.args.ctx)
			assert.NoError(t, err)
			assert.NotNil(t, events)

			file, err := fsh.RequestFile(tt.args.ctx, &filesharing.Transfer{
				Request: *req,
			})
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, file)
		})
	}
}
