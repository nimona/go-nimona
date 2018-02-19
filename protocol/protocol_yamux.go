package protocol

import (
	"context"
	"errors"
	"strings"

	yamux "github.com/hashicorp/yamux"
	zap "go.uber.org/zap"

	address "github.com/nimona/go-nimona-fabric/address"
	conn "github.com/nimona/go-nimona-fabric/connection"
	logging "github.com/nimona/go-nimona-fabric/logging"
)

// YamuxProtocol is a multiplexer protocol based on yamux
type YamuxProtocol struct {
	sessions map[string]*yamux.Session
	handler  func(context.Context, conn.Conn) error
}

// NewYamux returns a new yamus protocol and transport
func NewYamux() *YamuxProtocol {
	return &YamuxProtocol{
		sessions: map[string]*yamux.Session{},
	}
}

// Name of the protocol
func (m *YamuxProtocol) Name() string {
	return "yamux"
}

// Handle is the protocol handler for the server
func (m *YamuxProtocol) Handle(fn HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c conn.Conn) error {
		lgr := logging.Logger(ctx)

		ses, err := yamux.Server(c, nil)
		if err != nil {
			return err
		}

		str, err := ses.Accept()
		if err != nil {
			return err
		}

		addr := c.GetAddress()
		addr.Pop()
		lgr.Info("Handle: Accepting yamux sessions")

		go m.accept(ctx, ses, addr, fn)

		nc := conn.NewConnWrapper(str, c.GetAddress())
		return fn(ctx, nc)
	}
}

func (m *YamuxProtocol) accept(ctx context.Context, session *yamux.Session, addr *address.Address, fn HandlerFunc) {
	lgr := logging.Logger(ctx)
	for {
		str, err := session.Accept()
		if err != nil {
			lgr.Debug("Handle: Could not accept steam", zap.Error(err))
			// TODO break?
			break
		}

		// TODO copy addr
		nc := conn.NewConnWrapper(str, addr)
		if err := fn(ctx, nc); err != nil {
			lgr.Debug("Handle: Could not handle stream", zap.Error(err))
			// TODO break?
			continue
		}
	}
}

// Negotiate handles the client's side of the yamux protocol
func (m *YamuxProtocol) Negotiate(fn NegotiatorFunc) NegotiatorFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c conn.Conn) error {
		lgr := logging.Logger(ctx)

		session, err := yamux.Client(c, nil)
		if err != nil {
			return err
		}

		str, err := session.Open()
		if err != nil {
			return err
		}

		addr := c.GetAddress()
		addr.Pop()
		sessionAddr := strings.Join(addr.Processed(), "/")
		lgr.Info("Negotiage: Storing yamux session", zap.String("address", sessionAddr))
		m.sessions[sessionAddr] = session

		nc := conn.NewConnWrapper(str, c.GetAddress())
		return fn(ctx, nc)
	}
}

// CanDial checks if we can dial this address, and if so return the part of
// the address that will be consumed.
// This will only return true if the connection has been previously
// established and the connection is still open.
func (m *YamuxProtocol) CanDial(addr *address.Address) (bool, error) {
	as := addr.String()
	for k := range m.sessions {
		if strings.HasPrefix(as, k) {
			return true, nil
		}
	}

	return false, nil
}

// DialContext dials an address, assuming we have previously connected and
// negotiated yamux.
func (m *YamuxProtocol) DialContext(ctx context.Context, addr *address.Address) (context.Context, conn.Conn, error) {
	lgr := logging.Logger(ctx)
	lgr.Info("DialContext with yamux", zap.String("address", addr.String()))
	for k, ses := range m.sessions {
		if strings.HasPrefix(addr.String(), k) {
			lgr.Info("Found yamux session", zap.String("address", k))
			str, err := ses.Open()
			if err != nil {
				return nil, nil, err
			}

			parts := strings.Split(k, "/")
			for i := 0; i < len(parts); i++ {
				addr.Pop()
			}
			nc := conn.NewConnWrapper(str, addr)
			return ctx, nc, nil
		}
	}

	return nil, nil, errors.New("No such connection already open")
}

// Addresses returns the addresses the transport is listening to
func (m *YamuxProtocol) Addresses() []string {
	return []string{}
}

// Listen handles the transports
func (m *YamuxProtocol) Listen(ctx context.Context, handler HandlerFunc) error {
	m.handler = handler
	return nil
}
