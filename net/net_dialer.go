package net

import (
	"context"
	"errors"
	"reflect"

	zap "go.uber.org/zap"
)

var (
	// ErrCouldNotDial when no transports are available or internal error occured
	ErrCouldNotDial = errors.New("Could not dial")
)

// RequestIDKey for context
type RequestIDKey struct{}

// DialContext will attempt to connect to the given address and go through the
// various middlware that it needs until the connection is fully established
func (f *nnet) DialContext(ctx context.Context, as string) (context.Context, Conn, error) {
	if val := ctx.Value(RequestIDKey{}); val != nil {
		ctx = context.WithValue(ctx, RequestIDKey{}, generateReqID())
	}
	lgr := Logger(ctx)
	lgr.Debug("Dialing", zap.String("address", as))

	// TODO validate the address
	addr := NewAddress(as)

	// find transport we can dial
	// TODO figure out priorities, eg yamux should be more important than tcp
	for _, tr := range f.transports {
		if cd, err := tr.Transport.CanDial(addr); !cd || err != nil {
			continue
		}
		// dial transport
		var err error
		trType := reflect.TypeOf(tr.Transport).String()
		lgr.Debug("Attempting to dial", zap.String("transport", trType))
		newCtx, newConn, err := tr.Transport.DialContext(ctx, addr)
		if err != nil {
			lgr.Warn("Could not dial", zap.String("transport", trType), zap.Error(err))
			continue
		}

		newAddr := newConn.GetAddress()

		lgr.Debug("Dial complete, negotiating",
			zap.String("address", newAddr.String()),
			zap.String("Remaining", newAddr.RemainingString()),
		)

		// create chain with remaining protocols
		remProtocols := make([]Protocol, len(newAddr.RemainingProtocols()))
		for i, prName := range newAddr.RemainingProtocols() {
			protocol, ok := f.protocols[prName]
			if !ok {
				lgr.Debug("No such protocol", zap.String("protocol", prName))
				return nil, nil, ErrInvalidProtocol
			}
			remProtocols[i] = protocol
		}

		var retCtx context.Context
		var retConn Conn
		retProtocol := &EmptyProtocol{
			Negotiator: func(ctx context.Context, c Conn) error {
				retCtx = ctx
				retConn = c
				return nil
			},
		}

		remProtocols = append(remProtocols, retProtocol)
		chain := NegotiatorChain(remProtocols...)
		if err := chain(newCtx, newConn); err != nil {
			lgr.Warn("Could not negotiate", zap.Error(err))
			return nil, nil, ErrCouldNotDial
		}

		return retCtx, retConn, nil
	}

	return nil, nil, ErrCouldNotDial
}
