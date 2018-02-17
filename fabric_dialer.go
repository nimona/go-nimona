package fabric

import (
	"context"
	"errors"
	"reflect"

	zap "go.uber.org/zap"

	address "github.com/nimona/go-nimona-fabric/address"
	logging "github.com/nimona/go-nimona-fabric/logging"
	protocol "github.com/nimona/go-nimona-fabric/protocol"
)

var (
	// ErrCouldNotDial when no transports are available or internal error occured
	ErrCouldNotDial = errors.New("Could not dial")
)

// RequestIDKey for context
type RequestIDKey struct{}

// DialContext will attempt to connect to the given address and go through the
// various middlware that it needs until the connection is fully established
func (f *Fabric) DialContext(ctx context.Context, as string) error {
	ctx = context.WithValue(ctx, RequestIDKey{}, generateReqID())
	lgr := logging.Logger(ctx)
	lgr.Info("Dialing", zap.String("address", as))

	// TODO validate the address
	addr := address.NewAddress(as)

	// find transport we can dial
	// TODO figure out priorities, eg yamux should be more important than tcp
	for _, tr := range f.transports {
		if cd, err := tr.Transport.CanDial(addr); !cd || err != nil {
			continue
		}
		// dial transport
		var err error
		trType := reflect.TypeOf(tr.Transport).String()
		lgr.Info("Attempting to dial", zap.String("transport", trType))
		newCtx, newConn, err := tr.Transport.DialContext(ctx, addr)
		if err != nil {
			lgr.Warn("Could not dial", zap.String("transport", trType), zap.Error(err))
			continue
		}

		newAddr := newConn.GetAddress()

		lgr.Info("Dial complete, negotiating",
			zap.String("address", newAddr.String()),
			zap.String("address.Remaining", newAddr.RemainingString()),
		)

		// create chain with remaining protocols
		remProtocols := make([]protocol.Protocol, len(newAddr.RemainingProtocols()))
		for i, prName := range newAddr.RemainingProtocols() {
			protocol, ok := f.protocols[prName]
			if !ok {
				lgr.Debug("No such protocol", zap.String("protocol", prName))
				return ErrInvalidProtocol
			}
			remProtocols[i] = protocol
		}
		chain := protocol.NegotiatorChain(remProtocols...)
		if err := chain(newCtx, newConn); err != nil {
			lgr.Warn("Could not negotiate", zap.String("transport", trType), zap.Error(err))
			continue
		}
		return nil
	}

	return ErrCouldNotDial
}
