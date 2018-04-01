package protocol

import (
	"context"
	"errors"

	zap "go.uber.org/zap"

	nnet "github.com/nimona/go-nimona/net"
)

// LocalIdentityKey for context
type LocalIdentityKey struct{}

// RemoteIdentityKey for context
type RemoteIdentityKey struct{}

// IdentityProtocol allows exchanging peer information
type IdentityProtocol struct {
	Local string
}

var (
	// ErrUnexpectedRemote when remote doesn't match the one we were expecting
	ErrUnexpectedRemote = errors.New("Unexpected remote")
)

// Name of the protocol
func (m *IdentityProtocol) Name() string {
	return "identity"
}

// Handle is the protocol handler for the server
func (m *IdentityProtocol) Handle(fn nnet.HandlerFunc) nnet.HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c nnet.Conn) error {
		ctx = context.WithValue(ctx, LocalIdentityKey{}, m.Local)

		lgr := nnet.Logger(ctx).With(
			zap.Namespace("identity"),
		)

		// client will tell us who they are
		remoteID, err := c.ReadToken()
		if err != nil {
			lgr.Warn("Could not read remote id", zap.Error(err))
			return err
		}
		lgr.Debug("Read remote id", zap.String("remote.id", string(remoteID)))

		// store client's identity
		ctx = context.WithValue(ctx, RemoteIdentityKey{}, string(remoteID))

		// tell client our identity
		if err := c.WriteToken([]byte(m.Local)); err != nil {
			lgr.Warn("Could not write local id", zap.Error(err))
			return err
		}
		lgr.Debug("Wrote local id")

		c.GetAddress().Pop()
		return fn(ctx, c)
	}
}

// Negotiate handles the client's side of the identity protocol
func (m *IdentityProtocol) Negotiate(fn nnet.NegotiatorFunc) nnet.NegotiatorFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c nnet.Conn) error {
		// store local identity to conn
		ctx = context.WithValue(ctx, LocalIdentityKey{}, m.Local)
		lgr := nnet.Logger(ctx).With(
			zap.Namespace("identity"),
		)

		// tell the server who we are
		if err := c.WriteToken([]byte(m.Local)); err != nil {
			lgr.Warn("Could not write local id", zap.Error(err))
			return err
		}

		// server should now respond with their identity
		remoteID, err := c.ReadToken()
		if err != nil {
			lgr.Warn("Could not read remote id", zap.Error(err))
			return err
		}
		lgr.Debug("Read remote id", zap.String("remote.id", string(remoteID)))

		// if an identity has been provided as the first address parameter then
		// we need to make sure that the other side matches.
		addr := c.GetAddress()
		if len(addr.CurrentParams()) > 0 {
			if addr.CurrentParams() != string(remoteID) {
				lgr.Warn("Unexpected remote id", zap.String("remote.id", string(remoteID)))
				return ErrUnexpectedRemote
			}
		}

		c.GetAddress().Pop()
		// store server's identity
		ctx = context.WithValue(ctx, RemoteIdentityKey{}, string(remoteID))

		return fn(ctx, c)
	}
}

func (s *IdentityProtocol) GetAddresses() []string {
	return []string{}
}
