package blob

import (
	"bufio"
	"io"
)

const maxCapacity = 1024 * 1024

func ToBlob(r io.Reader) (*Blob, error) {
	blob := Blob{}
	chunks := make([]*Chunk, 0)

	br := bufio.NewReaderSize(r, maxCapacity)
	for {
		buf := make([]byte, maxCapacity)
		_, err := br.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		ch := &Chunk{
			Data: buf,
		}

		chunks = append(chunks, ch)
	}

	blob.Chunks = chunks

	return &blob, nil
}
