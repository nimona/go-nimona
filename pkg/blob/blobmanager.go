package blob

import (
	"errors"

	"nimona.io/pkg/context"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/sqlobjectstore"
)

type (
	Requester interface {
		Request(ctx context.Context, hash object.Hash) (*Blob, error)
	}
	requester struct {
		store    *sqlobjectstore.Store
		resolver resolver.Resolver
		objmgr   objectmanager.ObjectManager
	}
	Option func(*requester)
)

func NewRequester(
	ctx context.Context,
	opts ...Option,
) Requester {
	rqr := &requester{}

	for _, opt := range opts {
		opt(rqr)
	}

	return rqr
}

func (r *requester) Request(
	ctx context.Context,
	hash object.Hash,
) (*Blob, error) {
	logger := log.
		FromContext(ctx).
		Named("blob").
		With(
			log.String("method", "blob.Request"),
		)

	// find peers
	peers, err := r.resolver.Lookup(ctx, resolver.LookupByContentHash(hash))
	if err != nil {
		return nil, err
	}

	if len(peers) == 0 {
		return nil, errors.New("no peers found")
	}

	// request the blob object excluding the nested chunks
	obj, err := r.objmgr.Request(
		ctx,
		hash,
		peers[0],
		true,
	)
	if err != nil {
		logger.Error("failed to retrieve blob", log.Error(err))
		return nil, err
	}

	chunksHash, err := getChunks(obj)
	if err != nil {
		return nil, err
	}

	chunks := []*Chunk{}

	blob := &Blob{}

	// Request all the chunks
	for _, ch := range chunksHash {
		chObj, err := r.objmgr.Request(
			ctx,
			ch,
			peers[0],
			true,
		)
		if err != nil {
			logger.Error("failed to request chunk", log.Error(err))

			return nil, err
		}

		chunk := &Chunk{}
		if err := chunk.FromObject(chObj); err != nil {
			logger.Error("failed to convert to chunk", log.Error(err))
			return nil, err
		}

		chunks = append(chunks, chunk)
	}
	blob.Chunks = chunks

	return blob, nil
}

// nolint: golint // stuttering is fine for this one
type BlobUnloaded struct {
	Metadata       object.Metadata `nimona:"metadata:m,omitempty"`
	ChunksUnloaded []object.Hash   `nimona:"chunks:ar,omitempty"`
}

func (e *BlobUnloaded) Type() string {
	return "nimona.io/Blob"
}

func (e BlobUnloaded) ToObject() *object.Object {
	o, err := object.Encode(&e)
	if err != nil {
		panic(err)
	}
	return o
}

func (e *BlobUnloaded) FromObject(o *object.Object) error {
	return object.Decode(o, e)
}

func getChunks(o *object.Object) ([]object.Hash, error) {
	b := &BlobUnloaded{}
	if err := object.Decode(o, b); err != nil {
		return nil, err
	}

	return b.ChunksUnloaded, nil
}
