package blob

import (
	"io"
	"os"
	"runtime"

	"github.com/docker/go-units"
	"github.com/gammazero/workerpool"

	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
	"nimona.io/pkg/log"
	"nimona.io/pkg/object"
	"nimona.io/pkg/objectmanager"
	"nimona.io/pkg/resolver"
	"nimona.io/pkg/tilde"
)

type (
	Requester interface {
		Request(
			ctx context.Context,
			hash tilde.Digest,
		) (*Blob, []*Chunk, error)
	}
	Manager interface {
		Requester
		ImportFromFile(
			ctx context.Context,
			inputPath string,
		) (*Blob, error)
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
) (*Blob, error) {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}

	// keep a list of all chunk hashes
	chunkHashes := []tilde.Digest{}

	// start a workerpool to store chunks
	wp := workerpool.New(r.importWorkers)
	chunksErr := make(chan error)
	store := func(chunk *object.Object) func() {
		return func() {
			if err := r.objectmanager.Put(ctx, chunk); err != nil {
				chunksErr <- err
				return
			}
		}
	}

	// go through the file, make chunks, and store them
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
		chunkObj, err := object.Marshal(chunk)
		if err != nil {
			return nil, err
		}
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
	blob := &Blob{
		Chunks: chunkHashes,
	}
	blobObj, err := object.Marshal(blob)
	if err != nil {
		return nil, err
	}
	// store it
	if err := r.objectmanager.Put(ctx, blobObj); err != nil {
		return nil, err
	}
	// and return its hash
	return blob, nil
}

func (r *manager) Request(
	ctx context.Context,
	hash tilde.Digest,
) (*Blob, []*Chunk, error) {
	logger := log.
		FromContext(ctx).
		Named("blob").
		With(
			log.String("method", "blob.Request"),
		)

	// find peers
	peers, err := r.resolver.LookupByContent(ctx, hash)
	if err != nil {
		return nil, nil, err
	}

	if len(peers) == 0 {
		return nil, nil, errors.Error("no peers found")
	}

	// request the blob object excluding the nested chunks
	obj, err := r.objectmanager.Request(
		ctx,
		hash,
		peers[0].Metadata.Owner,
	)
	if err != nil {
		logger.Error("failed to retrieve blob", log.Error(err))
		return nil, nil, err
	}

	chunksHashes, err := getChunks(obj)
	if err != nil {
		return nil, nil, err
	}

	chunks := []*Chunk{}

	blob := &Blob{}
	if err := object.Unmarshal(obj, blob); err != nil {
		return nil, nil, err
	}

	// Request all the chunks
	for _, ch := range chunksHashes {
		chObj, err := r.objectmanager.Request(
			ctx,
			ch,
			peers[0].Metadata.Owner,
		)
		if err != nil {
			logger.Error("failed to request chunk", log.Error(err))

			return nil, nil, err
		}

		chunk := &Chunk{}
		if err := object.Unmarshal(chObj, chunk); err != nil {
			logger.Error("failed to convert to chunk", log.Error(err))
			return nil, nil, err
		}

		chunks = append(chunks, chunk)
	}

	return blob, chunks, nil
}

func getChunks(o *object.Object) ([]tilde.Digest, error) {
	b := &Blob{}
	if err := object.Unmarshal(o, b); err != nil {
		return nil, err
	}
	return b.Chunks, nil
}
