package fabric

import (
	"context"
)

type NegotiatorFunc func(ctx context.Context, conn Conn) (context.Context, Conn, error)
