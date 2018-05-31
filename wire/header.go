package wire

import (
	"bufio"
	"encoding/binary"
	"io"
)

// ReadMessage reads a wire message
func ReadMessage(r io.Reader) ([]byte, error) {
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

	return b[:n], nil
}

// WriteMessage writes a wire message
func WriteMessage(w io.Writer, bs []byte) error {
	bw := bufio.NewWriter(w)
	vb := make([]byte, 17)
	n := binary.PutUvarint(vb, uint64(len(bs)))
	wb := append(vb[:n], bs...)
	if _, err := bw.Write(wb); err != nil {
		return err
	}

	return bw.Flush()
}
