package fabric

import (
	"context"
)

type AddNegotiatorFunc func(ctx context.Context, conn Conn) (err error)
