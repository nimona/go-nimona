package fabric

import (
	"context"
	"errors"
)

// SelectMiddleware is the selector middleware
type SelectMiddleware struct {
	Handlers map[string]Handler
}

// Handle is the middleware handler for the server
func (m *SelectMiddleware) Handle(ctx context.Context, c Conn, addr Address) (context.Context, Conn, Address, error) {
	// pop self from address
	// ns := addr.Pop()
	// pr := strings.Split(ns, ":")[1]

	// we need to negotiate what they need from us
	// read the next token, which is the request for the next middleware
	prot, err := ReadToken(c)
	if err != nil {
		return nil, nil, addr, err
	}

	if err := WriteToken(c, prot); err != nil {
		return nil, nil, addr, err
	}

	return ctx, c, addr, nil
}

// Negotiate handles the client's side of the nimona middleware
func (m *SelectMiddleware) Negotiate(ctx context.Context, c Conn) (context.Context, Conn, error) {
	pr := "params"

	if err := WriteToken(c, []byte(pr)); err != nil {
		return ctx, nil, err
	}

	if err := m.verifyResponse(c, pr); err != nil {
		return ctx, nil, err
	}

	return ctx, c, nil
}

func (m *SelectMiddleware) verifyResponse(c Conn, pr string) error {
	resp, err := ReadToken(c)
	if err != nil {
		return err
	}

	if string(resp) != pr {
		return errors.New("Invalid selector response")
	}

	return nil
}
