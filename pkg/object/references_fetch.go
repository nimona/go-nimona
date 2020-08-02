package object

import (
	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
)

var (
	ErrDone    = errors.New("done")
	ErrTimeout = errors.New("timeout")
)

type (
	objectRefs struct {
		context context.Context
		request FetcherFunc
		next    chan objectRefOrErr
	}
	ReferencesResults interface {
		Next() (*Object, error)
	}
	objectRefOrErr struct {
		object *Object
		err    error
	}
)

// LoadReferences will look for references in the given object, request the
// referred objects using the requestHandler, and will return them in a lazy
// loaded result.
func FetchReferences(
	ctx context.Context,
	requestHandler FetcherFunc,
	objectHash Hash,
) (ReferencesResults, error) {
	next := make(chan objectRefOrErr)
	obj, err := requestHandler(ctx, objectHash)
	if err != nil {
		return nil, err
	}

	go func() {
		next <- objectRefOrErr{
			object: obj,
		}
		refs := GetReferences(*obj)
		for _, ref := range refs {
			refObj, err := requestHandler(
				ctx,
				ref,
			)
			next <- objectRefOrErr{
				object: refObj,
				err:    err,
			}
			if err != nil {
				return
			}
		}
		close(next)
	}()

	g := &objectRefs{
		context: ctx,
		request: requestHandler,
		next:    next,
	}

	return g, nil
}

func (g *objectRefs) Next() (*Object, error) {
	select {
	case n, ok := <-g.next:
		if !ok {
			return nil, ErrDone
		}
		return n.object, n.err
	case <-g.context.Done():
		return nil, ErrTimeout
	}
}
