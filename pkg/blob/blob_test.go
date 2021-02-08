package blob_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/docker/go-units"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/iotest"
	"nimona.io/pkg/blob"
	"nimona.io/pkg/object"
)

func Test_blobReader_Read(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{{
		name:   "should pass, 1b",
		length: 1,
	}, {
		name:   "should pass, 256b",
		length: 256,
	}, {
		name:   "should pass, 1Kb",
		length: 1 * units.KB,
	}, {
		name:   "should pass, 4096b",
		length: 4096,
	}, {
		name:   "should pass, 1Mb",
		length: units.MB,
	}, {
		name:   "should pass, 4Mb",
		length: 4 * units.MB,
	}, {
		name:   "should pass, 9.9Mb",
		length: 9.9 * units.MB,
	}, {
		name:   "should pass, 10Mb",
		length: 10 * units.MB,
	}, {
		name:   "should pass, 11.1Mb",
		length: 11.1 * units.MB,
	}, {
		name:   "should pass, 100Mb",
		length: 100 * units.MB,
	}, {
		name:   "should pass, 100.1Mb",
		length: 100.1 * units.MB,
	}, {
		name:   "should pass, 200.1Mb",
		length: 200.1 * units.MB,
	}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			fr := iotest.ZeroReader(tt.length)

			// read file into blob
			bl, ch, err := blob.NewBlob(fr)
			assert.NoError(t, err)
			assert.NotEmpty(t, bl.Chunks)
			assert.NotEmpty(t, ch)

			// checking if the generated chunks have the correct total length
			total := 0
			for _, ch := range ch {
				total += len(ch.Data)
				assert.NotEmpty(t, ch.Data)
			}
			require.Equal(t, tt.length, total)

			// write blob into file
			n, err := iotest.DrainReader(blob.NewReader(ch))
			assert.NoError(t, err)
			assert.Equal(t, int64(tt.length), n)
		})
	}
}

func TestBlob_Hash(t *testing.T) {
	c := &blob.Chunk{
		Data: []byte("foo"),
	}
	b := &blob.Blob{
		Chunks: []object.Hash{c.ToObject().Hash()},
	}
	u := &blob.BlobUnloaded{
		ChunksUnloaded: []object.Hash{
			c.ToObject().Hash(),
		},
	}

	bh := b.ToObject().Hash()
	uh := u.ToObject().Hash()
	assert.Equal(t, bh, uh)
}

func TestBlob_ResponseHash(t *testing.T) {
	c := &blob.Chunk{
		Data: []byte("foo"),
	}
	b := &blob.Blob{
		Chunks: []object.Hash{c.ToObject().Hash()},
	}
	r := &object.Response{
		RequestID: "foo",
		Object:    b.ToObject(),
	}
	s, err := json.Marshal(r.ToObject().ToMap())
	require.NoError(t, err)

	fmt.Println(string(s))

	m := object.Map{}
	err = json.Unmarshal(s, &m)
	require.NoError(t, err)
	o := object.FromMap(m)

	s, err = json.Marshal(o.ToMap())
	require.NoError(t, err)

	fmt.Println(string(s))

	fmt.Println("---")
	bh := r.ToObject().Hash()
	fmt.Println("---")
	uh := o.Hash()
	fmt.Println("---")

	assert.Equal(t, bh, uh)
}

func TestBlob_ToMap(t *testing.T) {
	b := &blob.Blob{
		Chunks: []object.Hash{
			blob.Chunk{
				Data: []byte("foo"),
			}.ToObject().Hash(),
		},
	}
	s, err := json.Marshal(b.ToObject().ToMap())
	require.NoError(t, err)
	fmt.Println(string(s))

	m := object.Map{}
	err = json.Unmarshal(s, &m)
	require.NoError(t, err)
	o := object.FromMap(m)

	s2, err := json.Marshal(o.ToMap())
	require.NoError(t, err)
	fmt.Println(string(s2))
	assert.Equal(t, s, s2)

	b2 := &blob.Blob{}
	err = b2.FromObject(o)
	require.NoError(t, err)

	require.Equal(t, b, b2)
}
