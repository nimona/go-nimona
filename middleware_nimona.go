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

func (m *NimonaMiddleware) CanHandle(addr string) bool {
	parts := addrSplit(addr)
	return parts[0][0] == NimonaKey
}

func (m *NimonaMiddleware) Negotiate(ctx context.Context, ucon Conn) error {
	protocol := "params"

	conn, err := ucon.GetRawConn()
	if err != nil {
		return err
	}

	// once connected we need to negotiate the second part, which is the is
	// an identity middleware.
	fmt.Println("Select: Writing protocol token")
	if err := WriteToken(conn, []byte(protocol)); err != nil {
		fmt.Println("Could not write identity token", err)
		return err
	}

	// server should now respond with an ok message
	fmt.Println("Select: Reading response")
	resp, err := ReadToken(conn)
	if err != nil {
		fmt.Println("Error reading ok response", err)
		return err
	}

	if string(resp) != protocol {
		return errors.New("Invalid selector response")
	}

	return nil
}

func (m *NimonaMiddleware) CanNegotiate(addr string) bool {
	parts := addrSplit(addr)
	return parts[0][0] == NimonaKey
}
