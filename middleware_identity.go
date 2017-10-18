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

func (m *IdentityMiddleware) Handle(ctx context.Context, conn Conn) (Conn, error) {
	// store local identity to conn
	conn.SetValue("identity_local", m.Local)

	// client will tell us who they are
	fmt.Println("Identity.Handle: Reading remote id")
	remoteID, err := ReadToken(conn)
	if err != nil {
		fmt.Println("Could not read remote clients's identity", err)
		return nil, err
	}

	// store client's identity
	conn.SetValue("identity_remote", remoteID)

	// client will tell us who they are looking for
	fmt.Println("Identity.Handle: Reading requested id")
	requestID, err := ReadToken(conn)
	if err != nil {
		fmt.Println("Could not read client's requested identity", err)
		return nil, err
	}

	// check if this is us
	if string(requestID) != m.Local {
		// TODO tell client this is not us
		fmt.Println("Identity.Handle: Requested identity does not match our local", string(requestID), m.Local)
		return nil, errors.New("No such identity")
	}

	// tell client our identity
	fmt.Println("Identity.Handle: Writing local id", m.Local)
	if err := WriteToken(conn, []byte(m.Local)); err != nil {
		fmt.Println("Could not write local id to client", err)
		return nil, err
	}

	// return connection as it was
	return conn, nil
}

func (m *IdentityMiddleware) CanHandle(addr string) bool {
	parts := addrSplit(addr)
	return parts[0][0] == IdentityKey
}

func (m *IdentityMiddleware) Negotiate(ctx context.Context, conn Conn) (Conn, error) {
	// store local identity to conn
	conn.SetValue("identity_local", m.Local)

	// check that context contains the address part that we need to extract
	// the remote id we are asking for
	prt := ctx.Value(ContextKeyAddressPart).(string)
	if prt == "" {
		return nil, errors.New("Missing address part")
	}

	// tell the server who we are
	fmt.Println("Identity.Negotiate: Writing local id", m.Local)
	if err := WriteToken(conn, []byte(m.Local)); err != nil {
		fmt.Println("Could not write local id to server", err)
		return nil, err
	}

	// tell the server who we are looking for
	reqID := strings.Split(prt, ":")[1]
	fmt.Println("Identity.Negotiate: Writing requested id", reqID)
	if err := WriteToken(conn, []byte(reqID)); err != nil {
		fmt.Println("Could not write request id to server", err)
		return nil, err
	}

	// server should now respond with their identity
	fmt.Println("Identity.Negotiate: Reading response")
	remoteID, err := ReadToken(conn)
	if err != nil {
		fmt.Println("Could not read remote server's identity", err)
		return nil, err
	}

	// store server's identity
	conn.SetValue("identity_remote", remoteID)

	// return connection as it was
	return conn, nil
}

func (m *IdentityMiddleware) CanNegotiate(addr string) bool {
	parts := addrSplit(addr)
	return parts[0][0] == IdentityKey
}
