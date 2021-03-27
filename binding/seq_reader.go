package main

import (
	"sync"

	"nimona.io/pkg/object"
)

func NewSequentialReader(readers ...object.ReadCloser) object.ReadCloser {
	return &sequentialReader{
		readers:   readers,
		curReader: 0,
		allDone:   false,
	}
}

type (
	sequentialReader struct {
		mutex     sync.Mutex
		readers   []object.ReadCloser
		curReader int
		allDone   bool
	}
)

func (r *sequentialReader) Read() (*object.Object, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.allDone {
		return nil, object.ErrReaderDone
	}
read:
	o, err := r.readers[r.curReader].Read()
	if err != nil {
		r.curReader++
		if r.curReader >= len(r.readers) {
			r.allDone = true
			return nil, object.ErrReaderDone
		}
		goto read
	}
	return o, nil
}

func (r *sequentialReader) Close() {
	for _, rd := range r.readers {
		rd.Close()
	}
}
