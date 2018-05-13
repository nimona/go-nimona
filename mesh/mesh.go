package mesh

import (
	"context"
	"net"
)

type Mesh interface {
	Dial(ctx context.Context, peerID string, protocol ...string) (net.Conn, error)
}
