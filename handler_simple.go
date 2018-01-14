package fabric

import (
	"context"
	"strings"
)

type simpleHandler struct {
	protocol string
	handler  HandlerFunc
}

func (h *simpleHandler) Wrap(f HandlerFunc) HandlerFunc {
	return func(ctx context.Context, ucon Conn) error {
		return h.handler(ctx, ucon)
	}
}

func (h *simpleHandler) CanHandle(addr string) bool {
	return strings.HasPrefix(addr, h.protocol)
}
