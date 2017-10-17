package fabric

import (
	"context"
	"fmt"
	"net"
	"strings"
)

const (
	TCPKey = "tcp"
)

type TCP struct{}

func (t *TCP) DialContext(ctx context.Context, addr string) (net.Conn, error) {
	prts := addrSplit(addr)
	tcpa := strings.Join(prts[0][1:], ":")
	tcon, err := net.Dial("tcp", tcpa)
	if err != nil {
		fmt.Println("Could not connect to server", err)
		return nil, err
	}

	return tcon, nil
}

func (t *TCP) CanDial(addr string) bool {
	parts := addrSplit(addr)
	return parts[0][0] == TCPKey
}
