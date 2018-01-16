package fabric

import (
	"context"
	"errors"
)

// NimonaMiddleware is the selector middleware
type NimonaMiddleware struct {
	Handlers map[string]HandlerFunc
}

// HandlerWrapper is the middleware handler for the server
func (m *NimonaMiddleware) HandlerWrapper(f HandlerFunc) HandlerFunc {
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

// Negotiate handles the client's side of the nimona middleware
func (m *NimonaMiddleware) Negotiate(ctx context.Context, c Conn) (context.Context, Conn, error) {
	pr := "params"

	if err := WriteToken(c, []byte(pr)); err != nil {
		return ctx, nil, err
	}

	if err := m.verifyResponse(c, pr); err != nil {
		return ctx, nil, err
	}

	return ctx, c, nil
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
