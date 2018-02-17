package connection

import (
	"bufio"
	"encoding/binary"
	"io"
)

// ReadToken that was written with WriteToken
func ReadToken(r io.Reader) ([]byte, error) {
	br := bufio.NewReader(r)
	l, err := binary.ReadUvarint(br)
	if err != nil {
		return nil, err
	}

	b := make([]byte, l)
	n, err := io.ReadFull(br, b)
	if err != nil {
		return nil, err
	}

	return b[:n-1], nil

}

// WriteToken to an io.Writer, appending a linebreak after it
func WriteToken(w io.Writer, bs []byte) error {
	bw := bufio.NewWriter(w)
	vb := make([]byte, 16)
	n := binary.PutUvarint(vb, uint64(len(bs)+1))
	wb := append(vb[:n], bs...)
	wb = append(wb, '\n')
	if _, err := w.Write(wb); err != nil {
		return err
	}

	return bw.Flush()
}
