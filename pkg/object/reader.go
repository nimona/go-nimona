package object

import (
	"nimona.io/pkg/context"
	"nimona.io/pkg/errors"
)

const (
	ErrReaderDone = errors.Error("reader done")
)

type (
	Reader interface {
		Read() (*Object, error)
	}
	ReadCloser interface {
		Read() (*Object, error)
		Close()
	}
	readCloser struct {
		ctx     context.Context
		objects <-chan *Object
		errors  <-chan error
		closer  chan<- struct{}
	}
)

func NewReadCloser(
	ctx context.Context,
	objects <-chan *Object,
	errs <-chan error,
	closer chan<- struct{},
) ReadCloser {
	r := &readCloser{
		ctx:     ctx,
		errors:  errs,
		objects: objects,
		closer:  closer,
	}
	return r
}

func (r *readCloser) Read() (*Object, error) {
	select {
	case next, ok := <-r.objects:
		if !ok {
			return nil, ErrReaderDone
		}
		if next == nil {
			return nil, ErrReaderDone
		}
		return next, nil
	case err, ok := <-r.errors:
		if !ok {
			return nil, ErrReaderDone
		}
		return nil, err
	case <-r.ctx.Done():
		return nil, ErrReaderDone
	}
}

func (r *readCloser) Close() {
	r.closer <- struct{}{}
}

// ReadAll is a helper method that
func ReadAll(r Reader) ([]*Object, error) {
	os := []*Object{}
	for {
		o, err := r.Read()
		if err == ErrReaderDone {
			return os, nil
		}
		if o == nil {
			return os, nil
		}
		if err != nil {
			return nil, err
		}
		os = append(os, o)
	}
}

// NewReadCloserFromObjects is mainly used for testing and mocks that return
// a Reader, or ReadCloser.
func NewReadCloserFromObjects(objects []Object) ReadCloser {
	objectChan := make(chan *Object)
	errorChan := make(chan error)
	closeChan := make(chan struct{})

	reader := NewReadCloser(
		context.TODO(),
		objectChan,
		errorChan,
		closeChan,
	)

	go func() {
		defer close(objectChan)
		defer close(errorChan)
		for i := range objects {
			objectChan <- &objects[i]
		}
	}()

	return reader
}
