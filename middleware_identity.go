package fabric

import (
	"context"
	"errors"
	"fmt"
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
		fmt.Println("Identity.Handle: Requested identity does not match our local")
		return nil, errors.New("No such identity")
	}

	// tell client our identity
	fmt.Println("Identity.Handle: Writing local id")
	if err := WriteToken(conn, []byte(m.Local)); err != nil {
		fmt.Println("Could not write local id to client", err)
		return nil, err
	}

	// return connection as it was
	return conn, nil
}

func (m *IdentityMiddleware) Negotiate(ctx context.Context, conn Conn, param string) (Conn, error) {
	// store local identity to conn
	conn.SetValue("identity_local", m.Local)

	// tell the server who we are
	fmt.Println("Identity.Negotiate: Writing local id")
	if err := WriteToken(conn, []byte(m.Local)); err != nil {
		fmt.Println("Could not write local id to server", err)
		return nil, err
	}

	// tell the server who we are looking for
	fmt.Println("Identity.Negotiate: Writing requested id")
	if err := WriteToken(conn, []byte(param)); err != nil {
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
