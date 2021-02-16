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
			cid object.CID,
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

	// keep a list of all chunk cids
	chunkCIDs := []object.CID{}

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
		// and add its cid to our list
		chunkCIDs = append(chunkCIDs, chunkObj.CID())
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
		Chunks: chunkCIDs,
	}
	blobObj := blob.ToObject()
	// store it
	if _, err := r.objectmanager.Put(ctx, blobObj); err != nil {
		return nil, err
	}
	// and return its cid
	return blob, nil
}

func (r *manager) Request(
	ctx context.Context,
	cid object.CID,
) (*Blob, []*Chunk, error) {
	logger := log.
		FromContext(ctx).
		Named("blob").
		With(
			log.String("method", "blob.Request"),
		)

	// find peers
	peers, err := r.resolver.Lookup(ctx, resolver.LookupByCID(cid))
	if err != nil {
		return nil, nil, err
	}

	if len(peers) == 0 {
		return nil, nil, errors.New("no peers found")
	}

	// request the blob object excluding the nested chunks
	obj, err := r.objectmanager.Request(
		ctx,
		cid,
		peers[0],
	)
	if err != nil {
		logger.Error("failed to retrieve blob", log.Error(err))
		return nil, nil, err
	}

	chunksCID, err := getChunks(obj)
	if err != nil {
		return nil, nil, err
	}

	chunks := []*Chunk{}

	blob := &Blob{}
	if err := blob.FromObject(obj); err != nil {
		return nil, nil, err
	}

	// Request all the chunks
	for _, ch := range chunksCID {
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

func getChunks(o *object.Object) ([]object.CID, error) {
	b := &Blob{}
	if err := b.FromObject(o); err != nil {
		return nil, err
	}
	return b.Chunks, nil
}
