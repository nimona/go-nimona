package object

import (
	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
)

var (
	ErrTimeout = errors.New("timeout")
)

func GetReferences(o *Object) []Hash {
	refs := []Hash{}
	Traverse(o.Data, func(k string, v interface{}) bool {
		if v == nil {
			return true
		}
		h, ok := v.(Hash)
		if !ok {
			return true
		}
		refs = append(refs, h)
		return true
	})
	return refs
}

// FetchWithReferences will look for references in the given object, request the
// primary object and all referred objects using the getter, and will
// return them in a lazy loaded result.
func FetchWithReferences(
	ctx context.Context,
	getter GetterFunc,
	objectHash Hash,
) (ReadCloser, error) {
	obj, err := getter(ctx, objectHash)
	if err != nil {
		return nil, err
	}

	objectChan := make(chan *Object)
	errorChan := make(chan error)
	closeChan := make(chan struct{})

	reader := NewReadCloser(
		ctx,
		objectChan,
		errorChan,
		closeChan,
	)

	go func() {
		defer close(objectChan)
		defer close(errorChan)
		select {
		case <-ctx.Done():
			return
		case <-closeChan:
			return
		case objectChan <- obj:
			// all good
		}
		refs := GetReferences(obj)
		for _, ref := range refs {
			refObj, err := getter(
				ctx,
				ref,
			)
			if err != nil {
				errorChan <- err
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-closeChan:
				return
			case objectChan <- refObj:
				// all good
			}
		}
	}()

	return reader, nil
}
