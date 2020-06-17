package blob

import (
	"bufio"
	"io"
)

const maxCapacity = 1000 * 1000

type Reader interface {
	Read(p []byte) (n int, err error)
}

type blobReader struct {
	total      int
	chunkIndex int
	dataIndex  int
	blob       *Blob
}

func ToBlob(r io.Reader) (*Blob, error) {
	blob := Blob{}
	chunks := make([]*Chunk, 0)

	br := bufio.NewReaderSize(r, maxCapacity)
	for {
		buf := make([]byte, maxCapacity)
		n, err := br.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		ch := &Chunk{
			Data: buf[0:n],
		}

		chunks = append(chunks, ch)
	}

	blob.Chunks = chunks

	return &blob, nil
}
func FromBlob(bl *Blob) Reader {
	return &blobReader{
		blob:      bl,
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
		if bl.chunkIndex >= len(bl.blob.Chunks) {
			break
		}

		// check the remaining data on the current chunk
		remainingData := len(bl.blob.Chunks[bl.chunkIndex].Data[bl.dataIndex:])
		diff := remainingData - maxBuffer

		// find the limits for the current read
		lower := bl.dataIndex
		upper := maxBuffer + bl.dataIndex - dataRead

		if diff < 0 {
			upper = len(bl.blob.Chunks[bl.chunkIndex].Data)
		}

		// append to the temp buf
		tempBuf = append(tempBuf, bl.blob.Chunks[bl.chunkIndex].Data[lower:upper]...)
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
