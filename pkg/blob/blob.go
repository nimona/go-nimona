package blob

import (
	"bufio"
	"io"

	"nimona.io/pkg/object"
)

type Reader interface {
	Read(p []byte) (n int, err error)
}

type blobReader struct {
	total      int
	chunkIndex int
	dataIndex  int
	chunks     []*Chunk
}

func NewBlob(r io.Reader) (*Blob, []*Chunk, error) {
	blob := &Blob{}
	chunks := make([]*Chunk, 0)

	br := bufio.NewReaderSize(r, defaultChunkSize)
	for {
		buf := make([]byte, defaultChunkSize)
		n, err := br.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}

		ch := &Chunk{
			Data: buf[0:n],
		}

		chunks = append(chunks, ch)
		cho, err := object.Marshal(ch)
		if err != nil {
			return nil, nil, err
		}
		blob.Chunks = append(blob.Chunks, cho.Hash())
	}

	return blob, chunks, nil
}

func NewReader(chunks []*Chunk) Reader {
	return &blobReader{
		chunks:    chunks,
		dataIndex: 0,
	}
}

func (bl *blobReader) Read(p []byte) (n int, err error) {
	maxBuffer := len(p)
	dataRead := 0

	// we need to use a temp buf to combine slices from multiple reads
	tempBuf := make([]byte, 0)

	// read until while the buffer is full
	for dataRead < maxBuffer {
		// check if we have already read all the chunks
		if bl.chunkIndex >= len(bl.chunks) {
			break
		}

		// check the remaining data on the current chunk
		remainingData := len(bl.chunks[bl.chunkIndex].Data[bl.dataIndex:])
		diff := remainingData - maxBuffer

		// find the limits for the current read
		lower := bl.dataIndex
		upper := maxBuffer + bl.dataIndex - dataRead

		if diff < 0 {
			upper = len(bl.chunks[bl.chunkIndex].Data)
		}

		// append to the temp buf
		tempBuf = append(
			tempBuf,
			bl.chunks[bl.chunkIndex].Data[lower:upper]...)
		dataRead += upper - lower

		// adjust the indexes
		bl.dataIndex += upper - lower

		if diff < 0 {
			bl.dataIndex = 0
			bl.chunkIndex++
		}
	}

	bl.total += dataRead

	// finally copy to the right byte array
	if len(tempBuf) == 0 {
		return 0, io.EOF
	}

	copy(p, tempBuf)

	return dataRead, nil
}
