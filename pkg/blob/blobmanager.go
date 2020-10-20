package blob

import (
	"errors"
	"sync"
	"time"

	"nimona.io/pkg/context"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/peer"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/sqlobjectstore"
)

//go:generate $GOBIN/genny -in=$GENERATORS/syncmap_named/syncmap.go -out=requests_generated.go -pkg=blob gen "KeyType=string ValueType=request SyncmapName=requests"
const (
	peerLookupTime = 5
)

type (
	Requester interface {
		Request(ctx context.Context, hash object.Hash) (*Blob, error)
	}
	requester struct {
		store    *sqlobjectstore.Store
		resolver resolver.Resolver
		requests *RequestsMap
		objmgr   objectmanager.ObjectManager
	}
	request struct {
		peers []*peer.Peer
		mutex *sync.RWMutex
	}
	Option func(*requester)
)

func NewRequester(
	ctx context.Context,
	opts ...Option,
) Requester {
	rqr := &requester{
		requests: NewRequestsMap(),
	}

	for _, opt := range opts {
		opt(rqr)
	}

	return rqr
}

func (r *requester) Request(
	ctx context.Context,
	hash object.Hash,
) (*Blob, error) {
	req := &request{
		peers: make([]*peer.Peer, 0),
		mutex: &sync.RWMutex{},
	}
	logger := log.
		FromContext(ctx).
		Named("blob").
		With(
			log.String("method", "blob.Request"),
		)

	// find peers
	peersCh, err := r.resolver.Lookup(ctx, resolver.LookupByContentHash(hash))
	if err != nil {
		return nil, err
	}

	peerFound := &peer.Peer{}

	select {
	case peerFound = <-peersCh:
		req.mutex.Lock()
		req.peers = append(req.peers, peerFound)
		req.mutex.Unlock()
	case <-ctx.Done():
		return nil, errors.New("context")
	case <-time.After(peerLookupTime * time.Second):
		break
	}

	// request the blob object excluding the nested chunks
	obj, err := r.objmgr.Request(
		ctx,
		hash,
		peerFound,
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
			peerFound,
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

type blobUnloaded struct {
	Blob
	ChunksUnloaded []object.Hash `nimona:"chunks:ar,omitempty"`
}

func getChunks(o *object.Object) ([]object.Hash, error) {
	b := &blobUnloaded{}
	if err := object.Decode(o, b); err != nil {
		return nil, err
	}

	return b.ChunksUnloaded, nil
}
