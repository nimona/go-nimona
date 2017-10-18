package fabric

import (
	"context"
	"fmt"
	"net"
	"strings"
)

const (
	transportPrefixTCP = "tcp:"
)

// NewTransportTCP returns a new TCP transport
func NewTransportTCP() Transport {
	return &TCP{}
}

// TCP transport
type TCP struct{}

// DialContext attemps to dial to the peer with the given addr
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

// CanDial checks if we can dial to this addr
func (t *TCP) CanDial(addr string) bool {
	return strings.HasPrefix(addr, transportPrefixTCP)
}
