package fabric

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"time"
)

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

	if uint64(n) != l {
		return nil, errors.New("Invalid token length")
	}

	return b[:n-1], nil

}

func WriteToken(w io.Writer, bs []byte) error {
	time.Sleep(time.Millisecond * 2000)
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

func addrSplit(addr string) [][]string {
	parts := strings.Split(addr, "/")
	res := make([][]string, len(parts))
	for i, part := range parts {
		res[i] = strings.Split(part, ":")
	}
	return res
}

func addrJoin(res [][]string) string {
	parts := make([]string, len(res))
	for i, part := range res {
		parts[i] = strings.Join(part, ":")
	}
	return strings.Join(parts, "/")
}
