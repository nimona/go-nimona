package fabric

import (
	"context"
	"errors"
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
	return func(ctx context.Context, c Conn) error {
		// we need to negotiate what they need from us
		// read the next token, which is the request for the next middleware
		prot, err := ReadToken(c)
		if err != nil {
			return err
		}

		if err := WriteToken(c, prot); err != nil {
			return err
		}

		// TODO could/should this f(ctx, ucon)?
		hf := m.Handlers[string(prot)]

		return hf(ctx, c)
	}
}

func (m *NimonaMiddleware) Negotiate(ctx context.Context, c Conn) (Conn, error) {
	pr := "params"

	if err := WriteToken(c, []byte(pr)); err != nil {
		return nil, err
	}

	if err := m.verifyResponse(c, pr); err != nil {
		return nil, err
	}

	return c, nil
}

func (m *NimonaMiddleware) verifyResponse(c Conn, pr string) error {
	resp, err := ReadToken(c)
	if err != nil {
		return err
	}

	if string(resp) != pr {
		return errors.New("Invalid selector response")
	}

	return nil
}
