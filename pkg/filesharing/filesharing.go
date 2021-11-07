package filesharing

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"

	"nimona.io/internal/rand"
	"nimona.io/pkg/blob"
	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/log"
	"nimona.io/pkg/mesh"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/peer"
)

var ErrTransferRejected = errors.New("transfer rejected")

type (
	Filesharer interface {
		Listen(
			ctx context.Context,
		) (
			chan *Transfer,
			error,
		)
		RequestTransfer(
			ctx context.Context,
			file *File,
			peerKey crypto.PublicKey,
		) (string, error)
		RespondTransfer(
			ctx context.Context,
			transfer Transfer,
			accepted bool,
		) error
		RequestFile(
			ctx context.Context,
			transfer *Transfer,
		) (
			*os.File,
			error,
		)
	}
	fileSharer struct {
		objmgr         objectmanager.ObjectManager
		mesh           mesh.Mesh
		receivedFolder string
	}
	Transfer struct {
		Request TransferRequest
		Peer    crypto.PublicKey
	}
)

func New(
	objectManager objectmanager.ObjectManager,
	msh mesh.Mesh,
	receivedFolder string,
) Filesharer {
	return &fileSharer{
		objmgr:         objectManager,
		mesh:           msh,
		receivedFolder: receivedFolder,
	}
}

func (fsh *fileSharer) RequestTransfer(
	ctx context.Context,
	file *File,
	peerReq crypto.PublicKey,
) (string, error) {
	nonce := rand.String(8)
	req := &TransferRequest{
		File:  *file,
		Nonce: nonce,
	}

	ro, err := object.Marshal(req)
	if err != nil {
		return "", err
	}

	err = fsh.mesh.Send(
		ctx,
		ro,
		peerReq,
	)
	if err != nil {
		return "", err
	}

	return nonce, nil
}

func (fsh *fileSharer) Listen(
	ctx context.Context,
) (
	chan *Transfer,
	error,
) {
	logger := log.
		FromContext(ctx).
		Named("filesharing").
		With(
			log.String("method", "filesharing.Listen"),
		)

	reqs := make(chan *Transfer)

	go func() {
		subs := fsh.mesh.Subscribe()
		err := fsh.handleObjects(ctx, subs, reqs)
		if err != nil {
			logger.Error("error getting objects", log.Error(err))
			return
		}
	}()

	return reqs, nil
}

func (fsh *fileSharer) handleObjects(
	ctx context.Context,
	sub mesh.EnvelopeSubscription,
	reqs chan *Transfer,
) error {
	logger := log.
		FromContext(ctx).
		Named("filesharing").
		With(
			log.String("method", "filesharing.handleObjects"),
		)

	for {
		env, err := sub.Next()
		if err != nil {
			return err
		}

		switch env.Payload.Type {
		case TransferRequestType:
			req := &TransferRequest{}

			if err := object.Unmarshal(env.Payload, req); err != nil {
				logger.Error(
					"failed to load FileIntentRequest from payload",
					log.Error(err),
				)
				continue
			}
			trf := &Transfer{
				Peer:    env.Sender,
				Request: *req,
			}

			reqs <- trf
		case TransferResponseType:
			resp := &TransferResponse{}
			if err = object.Unmarshal(env.Payload, resp); err != nil {
				logger.Error("error loading from payload", log.Error(err))
				continue
			}

			if !resp.Accepted {
				continue
			}
		}
	}
}

func (fsh *fileSharer) RequestFile(
	ctx context.Context,
	transfer *Transfer,
) (
	*os.File,
	error,
) {
	// TODO this needs to be improved to not store data in memory
	// and do all the operations in disk
	chunks := []*blob.Chunk{}

	for _, ch := range transfer.Request.File.Chunks {
		chObj, err := fsh.objmgr.Request(
			ctx,
			ch,
			&peer.ConnectionInfo{
				PublicKey: transfer.Peer,
			},
		)
		if err != nil {
			return nil, err
		}

		chunk := &blob.Chunk{}
		if err := object.Unmarshal(chObj, chunk); err != nil {
			return nil, err
		}

		chunks = append(chunks, chunk)
	}

	_ = os.MkdirAll(fsh.receivedFolder, os.ModePerm)
	f, err := os.Create(filepath.Join(
		fsh.receivedFolder,
		transfer.Request.File.Name,
	))
	if err != nil {
		return nil, err
	}

	r := blob.NewReader(chunks)
	bf := bufio.NewReader(r)
	if _, err := io.Copy(f, bf); err != nil {
		return nil, err
	}

	done := &TransferDone{
		Nonce: transfer.Request.Nonce,
	}
	doneObj, err := object.Marshal(done)
	if err != nil {
		return nil, err
	}
	if err := fsh.mesh.Send(
		ctx,
		doneObj,
		transfer.Peer,
	); err != nil {
		return f, err
	}

	return f, nil
}

func (fsh *fileSharer) RespondTransfer(
	ctx context.Context,
	transfer Transfer,
	accepted bool,
) error {
	resp := &TransferResponse{
		Nonce:    transfer.Request.Nonce,
		Accepted: accepted,
	}
	ro, err := object.Marshal(resp)
	if err != nil {
		return err
	}
	err = fsh.mesh.Send(ctx, ro, transfer.Peer)
	if err != nil {
		return err
	}
	return nil
}
