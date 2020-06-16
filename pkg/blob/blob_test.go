package blob_test

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"testing"

	"nimona.io/pkg/blob"

	"github.com/stretchr/testify/assert"
)

func TestToBlob(t *testing.T) {
	tempFile := newTestFile()
	fr, err := os.Open(tempFile)
	assert.NoError(t, err)

	bl, err := blob.ToBlob(fr)
	assert.NoError(t, err)
	assert.NotEmpty(t, bl.Chunks)
	for _, ch := range bl.Chunks {
		assert.NotEmpty(t, ch.Data)
	}
}

func newTestFile() string {
	f, _ := ioutil.TempFile("", "blob.*.nimona")
	defer f.Close()

	data := make([]byte, 10*1024*1024)
	_, _ = rand.Read(data)
	_, _ = f.Write(data)

	return f.Name()
}
