package fabric

import (
	"context"
	"errors"
	"fmt"
)

const (
	NimonaKey = "nimona"
)

type NimonaMiddleware struct {
	Handlers map[string]HandlerFunc
}

func (m *NimonaMiddleware) Handle(name string, f HandlerFunc) error {
	m.Handlers[name] = f
	return nil
}

func (m *NimonaMiddleware) Wrap(f HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, ucon Conn) error {
		conn, err := ucon.GetRawConn()
		if err != nil {
			return err
		}

		// we need to negotiate what they need from us
		fmt.Println("HandleSelect: Reading protocol token")
		// read the next token, which is the request for the next middleware
		prot, err := ReadToken(conn)
		if err != nil {
			fmt.Println("Could not read token", err)
			return err
		}

		fmt.Println("HandleSelect: Read protocol token:", string(prot))
		fmt.Println("HandleSelect: Writing protocol as ack")

		if err := WriteToken(conn, prot); err != nil {
			fmt.Println("Could not write protocol ack", err)
			return err
		}

		// TODO could/should this f(ctx, ucon)?
		hf := m.Handlers[string(prot)]

		return hf(ctx, ucon)
	}
}

func (m *NimonaMiddleware) Negotiate(ctx context.Context, conn Conn) error {
	pr := "params"

	if err := m.sendRequest(conn, pr); err != nil {
		return err
	}

	return m.verifyResponse(conn, pr)
}

func (m *NimonaMiddleware) sendRequest(conn Conn, pr string) error {
	rcon, err := conn.GetRawConn()
	if err != nil {
		return err
	}

	return WriteToken(rcon, []byte(pr))
}

func (m *NimonaMiddleware) verifyResponse(conn Conn, pr string) error {
	rcon, err := conn.GetRawConn()
	if err != nil {
		return err
	}

	resp, err := ReadToken(rcon)
	if err != nil {
		return err
	}

	if string(resp) != pr {
		return errors.New("Invalid selector response")
	}

	return nil
}
