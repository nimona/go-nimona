package fabric

import (
	"context"
	"errors"
	"strings"

	"github.com/hashicorp/yamux"
)

// YamuxMiddleware is a multiplexer middleware based on yamux
type YamuxMiddleware struct {
	sessions map[string]*yamux.Session
}

// Handle is the middleware handler for the server
func (m *YamuxMiddleware) Handle(ctx context.Context, c Conn) (context.Context, Conn, error) {
	ses, err := yamux.Server(c, nil)
	if err != nil {
		return nil, nil, err
	}

	str, err := ses.Accept()
	if err != nil {
		return nil, nil, err
	}

	nc := newConnWrapper(str, c.GetAddress())
	return ctx, nc, nil
}

// Negotiate handles the client's side of the yamux middleware
func (m *YamuxMiddleware) Negotiate(ctx context.Context, c Conn) (context.Context, Conn, error) {

	session, err := yamux.Client(c, nil)
	if err != nil {
		return ctx, nil, err
	}

	str, err := session.Open()
	if err != nil {
		return ctx, nil, err
	}

	nc := newConnWrapper(str, c.GetAddress())
	return ctx, nc, nil
}

// CanDial checks if we can dial this address, and if so return the part of
// the address that will be consumed.
// This will only return true if the connection has been previously
// established and the connection is still open.
func (m *YamuxMiddleware) CanDial(addr Address) (bool, string, error) {
	as := addr.String()
	for k := range m.sessions {
		if strings.HasPrefix(as, k) {
			return true, k, nil
		}
	}

	return false, "", nil
}

// DialContext dials an address, assuming we have previously connected and
// negotiated yamux.
func (m *YamuxMiddleware) DialContext(ctx context.Context, addr string) (context.Context, Conn, error) {
	for k, ses := range m.sessions {
		if strings.HasPrefix(addr, k) {
			str, err := ses.Open()
			if err != nil {
				return nil, nil, err
			}

			// TODO fix address
			addr := NewAddress("")
			nc := newConnWrapper(str, &addr)
			return ctx, nc, nil
		}
	}

	return nil, nil, errors.New("No such connection already open")
}
