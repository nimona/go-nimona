package fabric

import (
	"context"

	"go.uber.org/zap"
)

// Negotiate will process the next protocol in the given address recursively
func (f *Fabric) Negotiate(curCtx context.Context, curConn Conn) (context.Context, Conn, error) {
	// keep our conn and ctx
	conn := curConn
	ctx := curCtx

	// go throught all the protocols
	for {
		addr := conn.GetAddress()
		if len(addr.Remaining()) == 0 {
			return ctx, conn, nil
		}

		// get protocol
		pr := addr.CurrentProtocol()
		lgr := Logger(ctx).With(zap.String("negotiator", pr))
		lgr.Debug("Negotiating next protocol.")

		// check if we have this negotiator
		// if we don't have it, just return to the user
		ng, ok := f.negotiators[pr]
		if !ok {
			lgr.Warn("Negotiator not found.")
			return ctx, conn, errNoMoreProtocols
		}

		// execute negotiator
		newCtx, newConn, err := ng(ctx, conn)
		if err != nil {
			return nil, nil, err
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
