package fabric

import (
	"context"
)

type HandlerFunc func(ctx context.Context, conn Conn) (err error)

type Wrapper func(HandlerFunc) HandlerFunc
