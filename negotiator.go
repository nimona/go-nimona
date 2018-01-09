package fabric

import (
	"context"
)

type Negotiator interface {
	Negotiate(ctx context.Context, conn Conn) (err error)
	CanNegotiate(addr string) bool
}
