package blob_test

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/blob"
)

func TestToBlob(t *testing.T) {
	tempFile := newTestFile(10)
	fr, err := os.Open(tempFile)
	assert.NoError(t, err)

	bl, err := blob.ToBlob(fr)
	assert.NoError(t, err)
	assert.NotEmpty(t, bl.Chunks)
	assert.Len(t, bl.Chunks, 10)
	for _, ch := range bl.Chunks {
		assert.NotEmpty(t, ch.Data)
	}
}

func BenchmarkToBlob1(b *testing.B) {
	tempFile := newTestFile(1)

	for n := 0; n < b.N; n++ {
		fr, _ := os.Open(tempFile)
		_, err := blob.ToBlob(fr)
		if err != nil {
			b.Fail()
		}
	}
}

func BenchmarkToBlob100(b *testing.B) {
	tempFile := newTestFile(100)

	for n := 0; n < b.N; n++ {
		fr, _ := os.Open(tempFile)
		_, err := blob.ToBlob(fr)
		if err != nil {
			b.Fail()
		}
	}
}

func BenchmarkToBlob1000(b *testing.B) {
	tempFile := newTestFile(1000)

	for n := 0; n < b.N; n++ {
		fr, _ := os.Open(tempFile)
		_, err := blob.ToBlob(fr)
		if err != nil {
			b.Fail()
		}
	}
}

// size in megabytes
func newTestFile(size int) string {
	return newTestFileBytes(size * 1000 * 1000)
}

// size in bytes
func newTestFileBytes(size int) string {
	f, _ := ioutil.TempFile("", "blob.*.nimona")
	defer f.Close()

	data := make([]byte, size)
	_, _ = rand.Read(data)
	_, _ = f.Write(data)

	return f.Name()
}

func TestFromBlob(t *testing.T) {
	tempFile := newTestFile(4)

	// get hash
	exphash := fileHash(t, tempFile)

	// read file into blob
	fr, err := os.Open(tempFile)
	assert.NoError(t, err)
	bl, err := blob.ToBlob(fr)
	assert.NoError(t, err)
	assert.NotEmpty(t, bl.Chunks)
	for _, ch := range bl.Chunks {
		assert.NotEmpty(t, ch.Data)
	}

	// create new empty file
	f, err := ioutil.TempFile("", "blob.*.nimona.new")
	require.NoError(t, err)

	// write blob into file
	br := bufio.NewReader(blob.FromBlob(bl))
	n, err := io.Copy(f, br)
	assert.NoError(t, err)
	f.Close()

	// get hash
	gothash := fileHash(t, f.Name())

	// check things
	assert.Equal(t, 4*1000*1000, int(n))
	assert.Equal(t, exphash, gothash)
}

func fileHash(t *testing.T, file string) string {
	hf, err := os.Open(file)
	defer hf.Close() // nolint
	require.NoError(t, err)
	h := sha256.New()
	_, err = io.Copy(h, hf)
	require.NoError(t, err)
	expHash := fmt.Sprintf("%x", h.Sum(nil))
	return expHash
}

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
		length: 1000,
	}, {
		name:   "should pass, 4096b",
		length: 4096,
	}, {
		name:   "should pass, 1Mb",
		length: 1000 * 1000,
	}, {
		name:   "should pass, 4Mb",
		length: 4 * 1000 * 1000,
	}, {
		name:   "should pass, 9.9Mb",
		length: 9.9 * 1000 * 1000,
	}, {
		name:   "should pass, 10Mb",
		length: 10 * 1000 * 1000,
	}, {
		name:   "should pass, 11.1Mb",
		length: 11.1 * 1000 * 1000,
	}, {
		name:   "should pass, 100Mb",
		length: 100 * 1000 * 1000,
	}, {
		name:   "should pass, 100.1Mb",
		length: 100.1 * 1000 * 1000,
	}, {
		name:   "should pass, 200.1Mb",
		length: 200.1 * 1000 * 1000,
	}}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tempFile := newTestFileBytes(tt.length)

			// get hash
			exphash := fileHash(t, tempFile)

			// read file into blob
			fr, err := os.Open(tempFile)
			assert.NoError(t, err)
			bl, err := blob.ToBlob(fr)
			assert.NoError(t, err)
			assert.NotEmpty(t, bl.Chunks)

			// checking if the generated chunks have the correct total length
			total := 0
			for _, ch := range bl.Chunks {
				total += len(ch.Data)
				assert.NotEmpty(t, ch.Data)
			}
			require.Equal(t, tt.length, total)

			// create new empty file
			f, err := ioutil.TempFile("", "blob.*.nimona.new")
			assert.NoError(t, err)

			// write blob into file
			br := bufio.NewReader(blob.FromBlob(bl))
			n, err := io.Copy(f, br)
			assert.NoError(t, err)
			f.Close()

			// get hash
			gothash := fileHash(t, f.Name())

			// check things
			assert.Equal(t, tt.length, int(n))
			assert.Equal(t, exphash, gothash)
		})
	}
}
