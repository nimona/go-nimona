package fabric

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	IdentityKey = "nimona"
)

type IdentityMiddleware struct {
	Local string
}

func (m *IdentityMiddleware) Wrap(f HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, conn Conn) error {
		conn.SetValue("identity_local", m.Local)

		// client will tell us who they are
		fmt.Println("Identity.Handle: Reading remote id")
		remoteID, err := ReadToken(conn)
		if err != nil {
			fmt.Println("Could not read remote clients's identity", err)
			return err
		}
		fmt.Println("Identity.Handle: Read remote id:", string(remoteID))

		// store client's identity
		conn.SetValue("identity_remote", string(remoteID))

		// client will tell us who they are looking for
		fmt.Println("Identity.Handle: Reading requested id")
		requestID, err := ReadToken(conn)
		if err != nil {
			fmt.Println("Could not read client's requested identity", err)
			return err
		}
		fmt.Println("Identity.Handle: Read requested id:", string(requestID))

		// check if this is us
		if string(requestID) != m.Local {
			// TODO tell client this is not us
			fmt.Println("Identity.Handle: Requested identity does not match our local", string(requestID), m.Local)
			return errors.New("No such identity")
		}

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

func (m *IdentityMiddleware) Negotiate(ctx context.Context, conn Conn) error {
	// store local identity to conn
	conn.SetValue("identity_local", m.Local)

	// check that context contains the address part that we need to extract
	// the remote id we are asking for
	prt := ctx.Value(ContextKeyAddressPart).(string)
	if prt == "" {
		return errors.New("Missing address part")
	}

	// tell the server who we are
	fmt.Println("Identity.Negotiate: Writing local id", m.Local)
	if err := WriteToken(conn, []byte(m.Local)); err != nil {
		fmt.Println("Could not write local id to server", err)
		return err
	}

	// tell the server who we are looking for
	reqID := strings.Split(prt, ":")[1]
	fmt.Println("Identity.Negotiate: Writing requested id", reqID)
	if err := WriteToken(conn, []byte(reqID)); err != nil {
		fmt.Println("Could not write request id to server", err)
		return err
	}

	// server should now respond with their identity
	fmt.Println("Identity.Negotiate: Reading response")
	remoteID, err := ReadToken(conn)
	if err != nil {
		fmt.Println("Could not read remote server's identity", err)
		return err
	}
	fmt.Println("Identity.Negotiate: Read response:", string(remoteID))

	// store server's identity
	conn.SetValue("identity_remote", remoteID)

	return nil
}
