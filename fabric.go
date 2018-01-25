package fabric

import (
	"context"
	"errors"
	"net"
	"strings"

	zap "go.uber.org/zap"
)

var (
	// ErrNoTransport for when there is no transport with which to dial the address
	ErrNoTransport = errors.New("Could not dial with available transports")
	// ErrInvalidMiddleware when our handler doesn't know about a middleware in the
	ErrInvalidMiddleware = errors.New("No such middleware")
	// errNoMoreProtocols when fabric cannot deal with any more
	errNoMoreProtocols = errors.New("No more protocols")
)

var (
	// ContextKeyRequestID attached to each request
	ContextKeyRequestID = contextKey("request_id")
)

// New instance of fabric
func New(ms ...Middleware) *Fabric {
	bms := make([]string, len(ms))
	for i, m := range ms {
		bms[i] = m.Name()
	}
	f := &Fabric{
		base:        bms,
		transports:  []Transport{},
		negotiators: map[string]NegotiatorFunc{},
		handlers:    map[string]HandlerFunc{},
	}
	for _, m := range ms {
		f.AddMiddleware(m)
	}
	return f
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
	ctx = context.WithValue(ctx, ContextKeyRequestID, generateReqID())
	lgr := Logger(ctx)
	lgr.Info("Dialing", zap.String("address", as))

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
			if err == errNoMoreProtocols {
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

// GetAddresses returns a list of addresses for all the current transports
func (f *Fabric) GetAddresses() []string {
	addresses := []string{}
	for _, tr := range f.transports {
		addresses = append(addresses, tr.Address())
	}

	return addresses
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

// Listen on all transports
func (f *Fabric) Listen(ctx context.Context) error {
	// TODO handle re-listening on fail
	// Iterate over available transports and start listening
	for _, t := range f.transports {
		Logger(ctx).Info("Listening on tranport.", zap.String("transport", t.Name()))
		if err := t.Listen(ctx, f.handleRequest); err != nil {
			return err
		}
	}
	return nil
}

// Handles incoming requests.
func (f *Fabric) handleRequest(ctx context.Context, tcon net.Conn) error {
	ctx = context.WithValue(ctx, ContextKeyRequestID, generateReqID())
	lgr := Logger(ctx)

	// wrap net.Conn in Conn
	addr := NewAddress(strings.Join(f.base, "/"))
	c := newConnWrapper(tcon, &addr)

	// close the connection when we're done
	defer c.Close()

	for {
		if len(c.GetAddress().Remaining()) == 0 {
			lgr.Debug("Ran out of address parts; breaking")
			break
		}

		pr := c.GetAddress().CurrentProtocol()
		lgr.Debug("Handling next middleware.", zap.String("protocol", pr))

		hf, ok := f.handlers[pr]
		if !ok {
			lgr.Warn("Could not find middleware.", zap.String("protocol", pr))
			return ErrInvalidMiddleware
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
		return ctx, c, errNoMoreProtocols
	}

	// get protocol
	pr := addr.CurrentProtocol()
	lgr := Logger(ctx).With(zap.String("middleware", pr))
	lgr.Debug("Negotiating next middleware.")

	// check if is negotiator
	// if we don't have it, just return to the user
	ng, ok := f.negotiators[pr]
	if !ok {
		lgr.Warn("Middleware not found.")
		return ctx, c, errNoMoreProtocols
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
		lgr.Info("Middleware returned an empty connection.")
		return nil, nil, errors.New("done")
	}

	// pop item from address
	nc.GetAddress().Pop()

	// and move on
	return nctx, nc, err
}
