package protocol

import (
	"context"
	"errors"
	"strings"

	zap "go.uber.org/zap"

	nnet "github.com/nimona/go-nimona/net"
)

var (
	// ErrNoSuchRoute when requests route does not exist
	ErrNoSuchRoute = errors.New("No such route")
	// ErrInvalidCommand when our router doesn't know about this command
	ErrInvalidCommand = errors.New("Invalid command")
)

// RouterProtocol is the selector protocol
type RouterProtocol struct {
	routes map[string][]nnet.Protocol
}

// NewRouter returns a new router protocol
func NewRouter() *RouterProtocol {
	return &RouterProtocol{
		routes: map[string][]nnet.Protocol{},
	}
}

// Name of the protocol
func (m *RouterProtocol) Name() string {
	return "router"
}

// Handle is the protocol handler for the server
func (m *RouterProtocol) Handle(fn nnet.HandlerFunc) nnet.HandlerFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c nnet.Conn) error {
		addr := c.GetAddress()
		lgr := nnet.Logger(ctx).With(
			zap.Namespace("protocol:router"),
			zap.String("addr.current", addr.Current()),
			zap.String("addr.params", addr.CurrentParams()),
		)
		lgr.Debug("Reading token")

		// we need to negotiate what they need from us
		// read the next token, which is the request for the next protocol
		pr, err := c.ReadToken()
		if err != nil {
			return err
		}
		lgr.Debug("Read token", zap.String("pr", string(pr)))

		pf := strings.Split(string(pr), " ")
		if len(pf) != 2 {
			return ErrInvalidCommand
		}

		cm := pf[0]
		pm := pf[1]

		c.GetAddress().Pop()
		switch cm {
		case "SEL":
			lgr.Debug("Handling SEL", zap.String("cm", cm), zap.String("pm", pm))
			return m.handleGet(ctx, c, pm)
		default:
			lgr.Debug("Invalid command", zap.String("cm", cm), zap.String("pm", pm))
			c.Close()
			return ErrInvalidCommand
		}
	}
}

func (m *RouterProtocol) handleGet(ctx context.Context, c nnet.Conn, remainingAddrString string) error {
	remainingAddr := strings.Split(remainingAddrString, "/")

	validRoute := ""
	for route := range m.routes {
		if strings.HasPrefix(remainingAddrString, route) {
			validRoute = route
			break
		}
	}

	if validRoute == "" {
		return ErrNoSuchRoute
	}

	// TODO not sure about append, might wanna cut the stack up to our index
	// and the append the new stack
	c.GetAddress().Append(remainingAddr...)

	if err := c.WriteToken([]byte("ACK " + remainingAddrString)); err != nil {
		return err
	}

	chain := nnet.HandlerChain(m.routes[validRoute]...)
	return chain(ctx, c)
}

// Negotiate handles the client's side of the nimona protocol
func (m *RouterProtocol) Negotiate(fn nnet.NegotiatorFunc) nnet.NegotiatorFunc {
	// one time scope setup area for middleware
	return func(ctx context.Context, c nnet.Conn) error {
		c.GetAddress().Pop()
		pr := c.GetAddress().RemainingString()
		if err := c.WriteToken([]byte("SEL " + pr)); err != nil {
			return err
		}

		resp, err := c.ReadToken()
		if err != nil {
			return err
		}

		if string(resp) != "ACK "+pr {
			return errors.New("Invalid selector response")
		}

		return fn(ctx, c)
	}
}

// AddRoute adds an allowed route made up of protocols
func (m *RouterProtocol) AddRoute(protocols ...nnet.Protocol) error {
	protocolNames := []string{}
	for _, protocol := range protocols {
		protocolNames = append(protocolNames, protocol.Name())
	}
	routeName := strings.Join(protocolNames, "/")
	m.routes[routeName] = protocols
	return nil
}
