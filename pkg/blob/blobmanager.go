package blob

import (
	"errors"
	"io"
	"os"
	"runtime"

	"github.com/docker/go-units"
	"github.com/gammazero/workerpool"

	"nimona.io/pkg/context"
	"nimona.io/pkg/hyperspace/resolver"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
)

type (
	Requester interface {
		Request(
			ctx context.Context,
			hash object.Hash,
		) (*Blob, []*Chunk, error)
	}
	Manager interface {
		Requester
		ImportFromFile(
			ctx context.Context,
			inputPath string,
		) (*BlobUnloaded, error)
	}
	manager struct {
		resolver      resolver.Resolver
		objectmanager objectmanager.ObjectManager
		chunkSize     int
		importWorkers int
	}
	Option func(*manager)
)

var (
	defaultImportWorkers = runtime.NumCPU()
	defaultChunkSize     = 1 * units.MB
)

func NewManager(
	ctx context.Context,
	opts ...Option,
) Manager {
	mgr := &manager{
		chunkSize:     defaultChunkSize,
		importWorkers: defaultImportWorkers,
	}
	for _, opt := range opts {
		opt(mgr)
	}
	return mgr
}

func (r *manager) ImportFromFile(
	ctx context.Context,
	inputPath string,
) (*BlobUnloaded, error) {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}

	// keep a list of all chunk hashes
	chunkHashes := []object.Hash{}

	// start a workerpool to store chunks
	wp := workerpool.New(r.importWorkers)
	chunksErr := make(chan error)
	store := func(chunk *object.Object) func() {
		return func() {
			if _, err := r.objectmanager.Put(ctx, chunk); err != nil {
				chunksErr <- err
				return
			}
		}
	}

	// go through the file, makee chunks, and store them
	for {
		chunkBody := make([]byte, r.chunkSize)
		n, err := inputFile.Read(chunkBody)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF {
			if n != 0 {
				panic("n != 0")
			}
			break
		}
		// construct the next chunk
		chunk := &Chunk{
			Data: chunkBody[:n],
		}
		chunkObj := chunk.ToObject()
		// store it
		wp.Submit(store(chunkObj))
		// and add its hash to our list
		chunkHashes = append(chunkHashes, chunkObj.Hash())
	}

	wp.StopWait()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-chunksErr:
		return nil, err
	default:
	}

	// and finally construct the blob with the gathered list of chunks
	blob := &BlobUnloaded{
		ChunksUnloaded: chunkHashes,
	}
	blobObj := blob.ToObject()
	// store it
	if _, err := r.objectmanager.Put(ctx, blobObj); err != nil {
		return nil, err
	}
	// and return its hash
	return blob, nil
}

func (r *manager) Request(
	ctx context.Context,
	hash object.Hash,
) (*Blob, []*Chunk, error) {
	logger := log.
		FromContext(ctx).
		Named("blob").
		With(
			log.String("method", "blob.Request"),
		)

	// find peers
	peers, err := r.resolver.Lookup(ctx, resolver.LookupByContentHash(hash))
	if err != nil {
		return nil, nil, err
	}

	if len(peers) == 0 {
		return nil, nil, errors.New("no peers found")
	}

	// request the blob object excluding the nested chunks
	obj, err := r.objectmanager.Request(
		ctx,
		hash,
		peers[0],
	)
	if err != nil {
		logger.Error("failed to retrieve blob", log.Error(err))
		return nil, nil, err
	}

	chunksHash, err := getChunks(obj)
	if err != nil {
		return nil, nil, err
	}

	chunks := []*Chunk{}

	blob := &Blob{}
	if err := blob.FromObject(obj); err != nil {
		return nil, nil, err
	}

	// Request all the chunks
	for _, ch := range chunksHash {
		chObj, err := r.objectmanager.Request(
			ctx,
			ch,
			peers[0],
		)
		if err != nil {
			logger.Error("failed to request chunk", log.Error(err))

			return nil, nil, err
		}

		chunk := &Chunk{}
		if err := chunk.FromObject(chObj); err != nil {
			logger.Error("failed to convert to chunk", log.Error(err))
			return nil, nil, err
		}

		chunks = append(chunks, chunk)
	}

	return blob, chunks, nil
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
	c := make(object.StringArray, len(e.ChunksUnloaded))
	for i, v := range e.ChunksUnloaded {
		c[i] = object.String(v)
	}
	o := &object.Object{
		Type:     e.Type(),
		Metadata: e.Metadata,
		Data: object.Map{
			"chunks": c,
		},
	}
	return o
}

func (e *BlobUnloaded) FromObject(o *object.Object) error {
	e.Metadata = o.Metadata
	if v, ok := o.Data["chunks"]; ok {
		if t, ok := v.(object.StringArray); ok {
			c := make([]object.Hash, len(t))
			for i, v := range t {
				c[i] = object.Hash(v)
			}
			e.ChunksUnloaded = c
		}
	}
	return nil
}

func getChunks(o *object.Object) ([]object.Hash, error) {
	b := &BlobUnloaded{}
	if err := b.FromObject(o); err != nil {
		return nil, err
	}
	return b.ChunksUnloaded, nil
}
