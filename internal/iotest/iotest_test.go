package iotest

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZeroReader(t *testing.T) {
	r := ZeroReader(10)
	b, err := ioutil.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, b)
}

func TestDrainReader(t *testing.T) {
	r := ZeroReader(10)
	n, err := DrainReader(r)
	require.NoError(t, err)
	require.Equal(t, int64(10), n)
	p := make([]byte, 10)
	m, err := r.Read(p)
	require.Equal(t, io.EOF, err)
	assert.Equal(t, 0, m)
}
