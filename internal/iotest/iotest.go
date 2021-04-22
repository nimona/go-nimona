package iotest

import (
	"io"
	"io/ioutil"
)

// ZeroReader creates a reader that will return n zeroes
func ZeroReader(n int) io.Reader {
	return &repeatReader{
		n: n,
	}
}

type repeatReader struct {
	n int
	r int
}

func (r *repeatReader) Read(p []byte) (int, error) {
	if r.r >= r.n {
		return 0, io.EOF
	}
	l := r.n - r.r
	if l > len(p) {
		r.r += len(p)
		return len(p), nil
	}
	r.r = r.n
	return l, nil
}

func DrainReader(r io.Reader) (int64, error) {
	return io.Copy(ioutil.Discard, r)
}
