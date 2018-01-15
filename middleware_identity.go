package fabric

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

func (m *IdentityMiddleware) HandlerWrapper(f HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, conn Conn) error {
		ctx = context.WithValue(ctx, ContextKeyLocalIdentity, m.Local)

		// client will tell us who they are
		fmt.Println("Identity.Handle: Reading remote id")
		remoteID, err := ReadToken(conn)
		if err != nil {
			fmt.Println("Could not read remote clients's identity", err)
			return err
		}
		fmt.Println("Identity.Handle: Read remote id:", string(remoteID))

		// store client's identity
		ctx = context.WithValue(ctx, ContextKeyRemoteIdentity, string(remoteID))

		// tell client our identity
		fmt.Println("Identity.Handle: Writing local id", m.Local)
		if err := WriteToken(conn, []byte(m.Local)); err != nil {
			fmt.Println("Could not write local id to client", err)
			return err
		}
		fmt.Println("Identity.Handle: Wrote local id")

		return f(ctx, conn)
	}
}

func (m *IdentityMiddleware) Negotiate(ctx context.Context, conn Conn) (Conn, error) {
	// store local identity to conn
		ctx = context.WithValue(ctx, ContextKeyLocalIdentity, m.Local)

		// check that context contains the address part that we need to extract
		// the remote id we are asking for
		prt := "identity:SERVER" // TODO find a way to get current address part
		// prt := ctx.Value(ContextKeyAddressPart).(string)
		// if prt == "" {
		// 	return nil, errors.New("Missing address part")
		// }

		// tell the server who we are
		fmt.Println("Identity.NegotiatorWrapper: Writing local id", m.Local)
		if err := WriteToken(conn, []byte(m.Local)); err != nil {
			fmt.Println("Could not write local id to server", err)
			return nil, err
		}

		// server should now respond with their identity
		fmt.Println("Identity.NegotiatorWrapper: Reading response")
		remoteID, err := ReadToken(conn)
		if err != nil {
			fmt.Println("Could not read remote server's identity", err)
			return nil, err
		}
		fmt.Println("Identity.NegotiatorWrapper: Read response:", string(remoteID))

		exid := strings.Split(prt, ":")[1]
		if exid != string(remoteID) {
			return nil, errors.New("Unexpected remote server")
		}

		// store server's identity
		ctx = context.WithValue(ctx, ContextKeyRemoteIdentity, string(remoteID))

		return conn, nil
	}
}
