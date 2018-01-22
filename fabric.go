package fabric

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
)

var (
	ErrNoTransport      = errors.New("Could not dial with available transports")
	ErrNoSuchMiddleware = errors.New("No such middleware")
	ErrNoMoreProtocols  = errors.New("No more protocols")
)

// New instance of fabric
func New(ms ...Middleware) *Fabric {
	bms := make([]string, len(ms))
	for i, m := range ms {
		bms[i] = m.Name()
	}
	return &Fabric{
		base:        bms,
		transports:  []Transport{},
		negotiators: map[string]NegotiatorFunc{},
		handlers:    map[string]HandlerFunc{},
	}
}

// Fabric manages transports, negotiators, and handlers, and deals with Dialing.
type Fabric struct {
	base        []string
	transports  []Transport
	negotiators map[string]NegotiatorFunc
	handlers    map[string]HandlerFunc
}

// AddTransport for dialing to the outside world
func (f *Fabric) AddTransport(tr Transport) error {
	f.transports = append(f.transports, tr)
	return nil
}

// AddMiddleware for both client and server
func (f *Fabric) AddMiddleware(m Middleware) error {
	if err := f.AddHandlerFunc(m.Name(), m.Handle); err != nil {
		return err
	}
	return f.AddNegotiatorFunc(m.Name(), m.Negotiate)
}

// AddHandler for server
func (f *Fabric) AddHandler(m Handler) error {
	return f.AddHandlerFunc(m.Name(), m.Handle)
}

// AddNegotiator for client
func (f *Fabric) AddNegotiator(m Negotiator) error {
	return f.AddNegotiatorFunc(m.Name(), m.Negotiate)
}

// AddHandlerFunc for server
func (f *Fabric) AddHandlerFunc(r string, h HandlerFunc) error {
	f.handlers[r] = h
	return nil
}

// AddNegotiatorFunc for client
func (f *Fabric) AddNegotiatorFunc(n string, ng NegotiatorFunc) error {
	f.negotiators[n] = ng
	return nil
}

// DialContext will attempt to connect to the given address and go through the
// various middlware that it needs until the connection is fully established
func (f *Fabric) DialContext(ctx context.Context, as string) (context.Context, Conn, error) {
	// TODO validate the address
	addr := NewAddress(as)

	// figure out if the addr can be dialed and connect to the target
	c, err := f.dialTransport(ctx, addr)
	if err != nil {
		return ctx, nil, err
	}

	// pop first item which should be the transport
	c.GetAddress().Pop()

	// go throught all the protocols
	for {
		ctx, c, err = f.Next(ctx, c)
		if err != nil {
			if err == ErrNoMoreProtocols {
				return ctx, c, nil
			}
			return nil, nil, err
		}
	}
}

func (f *Fabric) dialTransport(ctx context.Context, addr Address) (Conn, error) {
	// get transport
	tr, err := f.getTransport(addr)
	if err != nil {
		return nil, ErrNoTransport
	}

	// dial
	tcon, err := tr.DialContext(ctx, addr)
	if err != nil {
		return nil, errors.New("Could not dial")
	}

	// create a new Conn that will be used to hold underlaying connections
	// from transports, middleware, as well as information about the
	// two parties.
	c := newConnWrapper(tcon, &addr)

	return c, nil
}

func (f *Fabric) getTransport(addr Address) (Transport, error) {
	// find transport we can dial
	// TODO figure out priorities, eg yamux should be more important than tcp
	for _, tr := range f.transports {
		cd, err := tr.CanDial(addr)
		if err != nil {
			return nil, err
		}
		if cd {
			return tr, nil
		}
	}

	return nil, ErrNoTransport
}

func (f *Fabric) Listen() error {
	// TODO replace with transport listens
	// TODO handle re-listening on fail

	// Iterate over available transports and start listening
	for _, t := range f.transports {
		err := t.Listen(f.handleRequest)
		if err != nil {
			return err
		}

	}
	return nil
}

// Handles incoming requests.
func (f *Fabric) handleRequest(tcon net.Conn) error {
	// wrap net.Conn in Conn
	addr := NewAddress(strings.Join(f.base, "/"))
	c := newConnWrapper(tcon, &addr)

	// close the connection when we're done
	defer c.Close()

	// TODO get earlier context
	ctx := context.Background()

	for {
		if len(c.GetAddress().Remaining()) == 0 {
			break
		}

		pr := c.GetAddress().CurrentProtocol()
		hf, ok := f.handlers[pr]
		if !ok {
			return ErrNoSuchMiddleware
		}

		var err error
		ctx, c, err = hf(ctx, c)
		if err != nil {
			return err
		}

		// TODO this is a weird check because of the ping handler that gets
		// executed and returns nil instead of a conn, not sure what to do
		// instead
		if c == nil {
			return nil
		}

		c.GetAddress().Pop()
	}

	return nil
}

// Next will process the next middleware in the given address recursively
func (f *Fabric) Next(ctx context.Context, c Conn) (context.Context, Conn, error) {
	addr := c.GetAddress()
	if len(addr.Remaining()) == 0 {
		return ctx, c, ErrNoMoreProtocols
	}

	// get protocol
	pr := addr.CurrentProtocol()
	fmt.Println("f.Next: pr=", pr)

	// check if is negotiator
	// if we don't have it, just return to the user
	ng, ok := f.negotiators[pr]
	if !ok {
		fmt.Println("f.Next: pr=", pr, "not found")
		return ctx, c, ErrNoMoreProtocols
	}

	// execute negotiator
	nctx, nc, err := ng(ctx, c)
	if err != nil {
		return nil, nil, err
	}

	// TODO this is a weird check because of the ping handler that gets
	// executed and returns nil instead of a conn, not sure what to do
	// instead
	if nc == nil {
		return nil, nil, errors.New("done")
	}

	// pop item from address
	nc.GetAddress().Pop()

	// and move on
	return nctx, nc, err
}
