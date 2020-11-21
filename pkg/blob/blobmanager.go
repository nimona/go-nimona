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
		) (*Blob, error)
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
	defaultChunkSize     = 256 * units.KB
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
	obj, err := r.objectmanager.Request(
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
	if err := blob.FromObject(obj); err != nil {
		return nil, err
	}

	// Request all the chunks
	for _, ch := range chunksHash {
		chObj, err := r.objectmanager.Request(
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
