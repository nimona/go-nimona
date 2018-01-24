package fabric

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"
)

var (
	// ContextKeyLocalIdentity is the key of the local identity in contexts
	ContextKeyLocalIdentity = contextKey("local_identity")
	// ContextKeyRemoteIdentity is the key of the remote identity in contexts
	ContextKeyRemoteIdentity = contextKey("remote_identity")
)

// IdentityMiddleware allows exchanging peer information
type IdentityMiddleware struct {
	Local string
}

// Name of the middleware
func (m *IdentityMiddleware) Name() string {
	return "identity"
}

// Handle is the middleware handler for the server
func (m *IdentityMiddleware) Handle(ctx context.Context, c Conn) (context.Context, Conn, error) {
	ctx = context.WithValue(ctx, ContextKeyLocalIdentity, m.Local)

	lgr := Logger(ctx).With(
		zap.Namespace("identity"),
	)

	// client will tell us who they are
	remoteID, err := ReadToken(c)
	if err != nil {
		lgr.Warn("Could not read remote id", zap.Error(err))
		return nil, nil, err
	}
	lgr.Debug("Read remote id", zap.String("remote.id", string(remoteID)))

	// store client's identity
	ctx = context.WithValue(ctx, ContextKeyRemoteIdentity, string(remoteID))

	// tell client our identity
	if err := WriteToken(c, []byte(m.Local)); err != nil {
		lgr.Warn("Could not write local id", zap.Error(err))
		return nil, nil, err
	}
	lgr.Debug("Wrote local id")

	return ctx, c, nil
}

// Negotiate handles the client's side of the identity middleware
func (m *IdentityMiddleware) Negotiate(ctx context.Context, conn Conn) (context.Context, Conn, error) {
	// store local identity to conn
	ctx = context.WithValue(ctx, ContextKeyLocalIdentity, m.Local)
	lgr := Logger(ctx).With(
		zap.Namespace("identity"),
	)

	// check that context contains the address part that we need to extract
	// the remote id we are asking for
	prt := "identity:SERVER" // TODO find a way to get current address part
	// prt := ctx.Value(ContextKeyAddressPart).(string)
	// if prt == "" {
	// 	return nil, errors.New("Missing address part")
	// }

	// tell the server who we are
	if err := WriteToken(conn, []byte(m.Local)); err != nil {
		lgr.Warn("Could not write local id", zap.Error(err))
		return ctx, nil, err
	}

	// server should now respond with their identity
	remoteID, err := ReadToken(conn)
	if err != nil {
		lgr.Warn("Could not read remote id", zap.Error(err))
		return ctx, nil, err
	}
	lgr.Info("Read remote id", zap.String("remote.id", string(remoteID)))

	exid := strings.Split(prt, ":")[1]
	if exid != string(remoteID) {
		lgr.Warn("Unexpected remote id", zap.String("remote.id", string(remoteID)))
		return ctx, nil, errors.New("Unexpected remote server")
	}

	// store server's identity
	ctx = context.WithValue(ctx, ContextKeyRemoteIdentity, string(remoteID))

	return ctx, conn, nil
}
