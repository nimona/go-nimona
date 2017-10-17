package fabric

import (
	"context"
	"strings"
)

type simpleHandler struct {
	protocol string
	handler  HandlerFunc
}

func (h *simpleHandler) Handle(ctx context.Context, conn Conn) (newConn Conn, err error) {
	return h.handler(ctx, conn)
}
func (h *simpleHandler) CanHandle(addr string) bool {
	return strings.HasPrefix(addr, h.protocol)
}
