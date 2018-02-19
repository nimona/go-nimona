package protocol

import (
	"context"

	connection "github.com/nimona/go-nimona-fabric/connection"
)

type Dialer interface {
	DialContext(ctx context.Context, as string) (context.Context, connection.Conn, error)
}
