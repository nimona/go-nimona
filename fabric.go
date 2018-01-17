package fabric

import (
	"context"
	"errors"
	"fmt"
	"net"
)

var (
	ErrNoTransport      = errors.New("Could not dial with available transports")
	ErrNoSuchMiddleware = errors.New("No such middleware")
	ErrNoMoreProtocols  = errors.New("No more protocols")
)

func New() *Fabric {
	return &Fabric{
		transports:  []Transport{},
		negotiators: map[string]NegotiatorFunc{},
		handlers:    map[string]Handler{},
	}
}

// Fabric manages transports, negotiators, and handlers, and deals with Dialing.
type Fabric struct {
	transports  []Transport
	negotiators map[string]NegotiatorFunc
	handlers    map[string]Handler
}

// AddTransport for dialing to the outside world
func (f *Fabric) AddTransport(n string, tr Transport) error {
	f.transports = append(f.transports, tr)
	return nil
}

// AddHandler for server
func (f *Fabric) AddHandler(r string, h Handler) error {
	f.handlers[r] = h
	return nil
}

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

	// pop first itam
	c.GetAddress().Pop()

	// handshake
	if err := f.handshake(c); err != nil {
		return ctx, nil, err
	}

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

func (f *Fabric) handshake(c Conn) error {
	rs := c.GetAddress().RemainingString()
	fmt.Println("Handshake:", rs)
	return WriteToken(c, []byte(rs))
}

func (f *Fabric) Listen() error {
	// TODO replace with transport listens
	// TODO handle re-listening on fail
	// go func() {
	l, err := net.Listen("tcp", "0.0.0.0:3000")
	if err != nil {
		return err
	}

	defer l.Close()
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func(conn net.Conn) {
			defer func() {
				if err := conn.Close(); err != nil {
					fmt.Println("Could not close conn", err)
				}
			}()
			if err := f.handleRequest(conn); err != nil {
				fmt.Println("Listen: Could not handle request. error:", err)
			}
		}(conn)
	}
	// }()
	// return nil
}

// Handles incoming requests.
func (f *Fabric) handleRequest(tcon net.Conn) error {
	// a client initiated a connection
	fmt.Println("handleRequest: New incoming connection")

	saddr, err := ReadToken(tcon)
	if err != nil {
		return err
	}
	fmt.Println("Handshake:", string(saddr))

	// wrap net.Conn in Conn
	addr := NewAddress(string(saddr))
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
		fmt.Println("> Current:", pr)
		hf, ok := f.handlers[pr]
		if !ok {
			return ErrNoSuchMiddleware
		}

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
		return nil, nil, ErrNoMoreProtocols
	}

	// get protocol
	pr := addr.CurrentProtocol()

	// check if is negotiator
	ng, ok := f.negotiators[pr]
	if !ok {
		return nil, nil, ErrNoSuchMiddleware // TODO Switch to err no negotiator
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
