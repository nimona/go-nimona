package fabric

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

var (
	// ErrNoSuchRoute when requests route does not exist
	ErrNoSuchRoute = errors.New("No such route")
)

// RouterProtocol is the selector protocol
type RouterProtocol struct {
	Handlers map[string]Protocol
	routes   []string
}

// NewRouter returns a new router protocol
func NewRouter() *RouterProtocol {
	return &RouterProtocol{
		Handlers: map[string]Protocol{},
		routes:   []string{},
	}
}

// Name of the protocol
func (m *RouterProtocol) Name() string {
	return "router"
}

// Handle is the protocol handler for the server
func (m *RouterProtocol) Handle(fn HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c Conn) error {
		addr := c.GetAddress()
		lgr := Logger(ctx).With(
			zap.Namespace("protocol:router"),
			zap.String("addr.current", addr.Current()),
			zap.String("addr.params", addr.CurrentParams()),
		)
		lgr.Debug("Reading token")

		// we need to negotiate what they need from us
		// read the next token, which is the request for the next protocol
		pr, err := ReadToken(c)
		if err != nil {
			return err
		}
		lgr.Debug("Read token", zap.String("pr", string(pr)))

		pf := strings.Split(string(pr), " ")
		if len(pf) != 2 {
			return errors.New("invalid router command format")
		}

		cm := pf[0]
		pm := pf[1]

		switch cm {
		case "SEL":
			lgr.Debug("Handling SEL", zap.String("cm", cm), zap.String("pm", pm))
			ctx, conn, err := m.handleGet(ctx, c, pm)
			if err != nil {
				return err
			}
			return fn(ctx, conn)
		default:
			lgr.Debug("Invalid command", zap.String("cm", cm), zap.String("pm", pm))
			c.Close()
			return errors.New("invalid router command")
		}
	}
}

func (m *RouterProtocol) handleGet(ctx context.Context, c Conn, pm string) (context.Context, Conn, error) {
	addr := c.GetAddress()

	remainingAddr := strings.Split(pm, "/")[1:]
	remainingAddrString := strings.Join(remainingAddr, "/")
	validRoute := false
	for _, route := range m.routes {
		if strings.HasPrefix(route, remainingAddrString) {
			validRoute = true
			break
		}
	}

	if !validRoute {
		return ctx, c, ErrNoSuchRoute
	}

	// TODO not sure about append, might wanna cut the stack up to our index
	// and the append the new stack
	addr.stack = append(addr.stack, remainingAddr...)

	if err := WriteToken(c, []byte("ACK "+pm)); err != nil {
		return nil, nil, err
	}

	return ctx, c, nil
}

// Negotiate handles the client's side of the nimona protocol
func (m *RouterProtocol) Negotiate(fn HandlerFunc) HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c Conn) error {
		pr := c.GetAddress().RemainingString()
		fmt.Println("Router.Negotiate: pr=", pr)

		if err := WriteToken(c, []byte("SEL "+pr)); err != nil {
			return err
		}

		if err := m.verifyResponse(c, "ACK "+pr); err != nil {
			return err
		}

		return fn(ctx, c)
	}
}

func (m *RouterProtocol) verifyResponse(c Conn, pr string) error {
	resp, err := ReadToken(c)
	if err != nil {
		return err
	}

	if string(resp) != pr {
		return errors.New("Invalid selector response")
	}

	return nil
}

// AddRoute adds an allowed route made up of protocols
func (m *RouterProtocol) AddRoute(protocols ...Protocol) error {
	protocolNames := []string{}
	for _, protocol := range protocols {
		protocolNames = append(protocolNames, protocol.Name())
	}
	m.routes = append(m.routes, strings.Join(protocolNames, "/"))
	return nil
}
