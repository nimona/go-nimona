package fabric

import (
	"context"
	"errors"

	"go.uber.org/zap"
)

// LocalIdentityKey for context
type LocalIdentityKey struct{}

// RemoteIdentityKey for context
type RemoteIdentityKey struct{}

// IdentityProtocol allows exchanging peer information
type IdentityProtocol struct {
	Local string
}

// Name of the protocol
func (m *IdentityProtocol) Name() string {
	return "identity"
}

// Handle is the protocol handler for the server
func (m *IdentityProtocol) Handle(fn HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c Conn) error {
		ctx = context.WithValue(ctx, LocalIdentityKey{}, m.Local)

		lgr := Logger(ctx).With(
			zap.Namespace("identity"),
		)

		// client will tell us who they are
		remoteID, err := ReadToken(c)
		if err != nil {
			lgr.Warn("Could not read remote id", zap.Error(err))
			return err
		}
		lgr.Debug("Read remote id", zap.String("remote.id", string(remoteID)))

		// store client's identity
		ctx = context.WithValue(ctx, RemoteIdentityKey{}, string(remoteID))

		// tell client our identity
		if err := WriteToken(c, []byte(m.Local)); err != nil {
			lgr.Warn("Could not write local id", zap.Error(err))
			return err
		}
		lgr.Debug("Wrote local id")

		c.GetAddress().Pop()
		return fn(ctx, c)
	}
}

// Negotiate handles the client's side of the identity protocol
func (m *IdentityProtocol) Negotiate(fn NegotiatorFunc) NegotiatorFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c Conn) error {
		// store local identity to conn
		ctx = context.WithValue(ctx, LocalIdentityKey{}, m.Local)
		lgr := Logger(ctx).With(
			zap.Namespace("identity"),
		)

		// tell the server who we are
		if err := WriteToken(c, []byte(m.Local)); err != nil {
			lgr.Warn("Could not write local id", zap.Error(err))
			return err
		}

		// server should now respond with their identity
		remoteID, err := ReadToken(c)
		if err != nil {
			lgr.Warn("Could not read remote id", zap.Error(err))
			return err
		}
		lgr.Info("Read remote id", zap.String("remote.id", string(remoteID)))

		// if an identity has been provided as the first address parameter then
		// we need to make sure that the other side matches.
		addr := c.GetAddress()
		if len(addr.CurrentParams()) > 0 {
			lgr.Warn("Unexpected remote id", zap.String("remote.id", string(remoteID)))
			return errors.New("Unexpected remote server")
		}

		c.GetAddress().Pop()
		// store server's identity
		ctx = context.WithValue(ctx, RemoteIdentityKey{}, string(remoteID))

		return fn(ctx, c)
	}
}
