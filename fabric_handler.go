package fabric

import (
	"context"
	"net"
	"strings"

	"go.uber.org/zap"
)

// Handle incoming requests.
func (f *Fabric) Handle(curCtx context.Context, curConn net.Conn) error {
	// wrap net.Conn in Conn
	curAddr := NewAddress(strings.Join(f.base, "/"))
	conn := newConnWrapper(curConn, &curAddr)
	ctx := context.WithValue(curCtx, ContextKeyRequestID, generateReqID())

	for {
		addr := conn.GetAddress()
		if len(addr.Remaining()) == 0 {
			return nil
		}

		// get protocol
		pr := addr.CurrentProtocol()
		lgr := Logger(ctx).With(zap.String("handler", pr))
		lgr.Debug("Handling next middleware.")

		// check if we have this handler
		// if we don't have it, just return to the user
		hf, ok := f.handlers[pr]
		if !ok {
			lgr.Warn("Handler not found.")
			return ErrInvalidMiddleware
		}

		// execute handler
		newCtx, newConn, err := hf(ctx, conn)
		if err != nil {
			return err
		}

		// if we got a new connection, replace conn with the new one
		if newConn != nil {
			conn = newConn
		}
		// same with context
		if newCtx != nil {
			ctx = newCtx
		}

		// pop item from address
		conn.GetAddress().Pop()
	}
}
