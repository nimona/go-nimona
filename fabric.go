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

	ContextKeyAddressPart = contextKey("addrpart")
)

func New() *Fabric {
	return &Fabric{
		transports:  map[string]Transport{},
		negotiators: map[string]Negotiator{},
		routes:      map[string]HandlerFunc{},
	}
}

type Fabric struct {
	// handlers    []Handler
	negotiators map[string]Negotiator
	transports  map[string]Transport
	routes      map[string]HandlerFunc
}

func (f *Fabric) AddTransport(n string, tr Transport) error {
	f.transports[n] = tr
	return nil
}

func (f *Fabric) AddHandlerFunc(r string, hf HandlerFunc) error {
	f.routes[r] = hf
	return nil
}

func (f *Fabric) AddNegotiator(n string, ng Negotiator) error {
	f.negotiators[n] = ng
	return nil
}

func (f *Fabric) DialContext(ctx context.Context, addr string) (Conn, error) {
	// TODO validate the address

	// get the stack of middleware that we need to go through
	// the connection should be considered successful once this array is empty
	stack := strings.Split(addr, "/")

	// TODO close connection when done if something errored
	// defer go func() { if .. conn.Close() }()

	// once we are connected to the transport part we need to negotiate and
	// go through the various middleware

	conn := &conn{
		fabric: f,
		values: map[string]interface{}{},
		stack:  stack,
	}

	// go through all the parts of the address that are left in the stack
	// TODO figure out how to deal with the last item in the stack
	// ^ this is a weird one.
	// let's assume the client is dialing `tcp:127.0.0.1:3000/nimona:SERVER/ping`
	// it will connect via TCP transport, will go through the `nimona` middlware,
	// and then the connection should probably be returned to the called so it
	// can use it.

	// get next protocol
	ns := conn.popStack()
	fmt.Println("Processing", ns)

	// get protocol
	pr := strings.Split(ns, ":")[0]

	// check if is transport
	tr, ok := f.transports[pr]
	if !ok {
		return nil, ErrNoTransport
	}

	tcon, err := tr.DialContext(ctx, ns)
	if err != nil {
		// TODO log and handle error correctly
		fmt.Println("Could not dial", ns, err)
		return nil, errors.New("Could not dial")
	}

	// upgrade our connection
	if err := conn.Upgrade(tcon); err != nil {
		return nil, err
	}

	// handshake
	hs := conn.remainingStackString() // TODO Add "STREAM" + ...
	if err := WriteToken(conn, []byte(hs)); err != nil {
		return nil, err
	}

	if err := f.Next(ctx, conn); err != nil {
		fmt.Println("Error on conn.Next()", err)
		return nil, err
	}

	return conn, nil
}

func (f *Fabric) Listen() error {
	// TODO replace with transport listens
	// TODO handle re-listening on fail
	// go func() {
	l, err := net.Listen("tcp", "0.0.0.0:3000")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return err
	}
	defer l.Close()
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
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

	// wrap net.Conn in Conn
	c := &conn{
		conn:   tcon,
		fabric: f,
		values: map[string]interface{}{},
		stack:  []string{},
	}

	// close the connection when we're done
	defer c.Close()

	rt, err := ReadToken(tcon)
	if err != nil {
		return err
	}

	fmt.Println("Got rt:", string(rt))

	hf, ok := f.routes[string(rt)]
	if !ok {
		return ErrNoSuchMiddleware // TODO Cannot find route or something
	}

	ctx := context.Background()
	return hf(ctx, c)

	// handle all middleware that we are being given
	// for {
	// 	fmt.Println("handleRequest: Waiting for next protocol")
	// 	// get next protocol name
	// 	prot, err := f.HandleSelect(c)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	fmt.Println("handleRequest: Got next protocol", prot)

	// 	// check if there is a middleware for this
	// 	hnd := false
	// 	for _, mid := range f.handlers {
	// 		if !mid.CanHandle(prot) {
	// 			continue
	// 		}
	// 		// execute middleware
	// 		if err := mid.Handle(context.Background(), c); err != nil {
	// 			// TODO should we be closing the connection here?
	// 			c.Close()
	// 			return err
	// 		}
	// 		// mark as handled
	// 		fmt.Println("handleRequest: Handled protocol", prot)
	// 		hnd = true
	// 		// and move on
	// 		break
	// 	}
	// 	if !hnd {
	// 		fmt.Println("handleRequest: Could not handle", prot)
	// 		return ErrNoSuchMiddleware
	// 	}
	// }

	// return nil
}

func (f *Fabric) Next(ctx context.Context, c *conn) error {
	// get next protocol
	ns := c.popStack()
	fmt.Println("Processing", ns)

	// get protocol
	pr := strings.Split(ns, ":")[0]

	// check if is negotiator
	ng, ok := f.negotiators[pr]
	if !ok {
		return ErrNoSuchMiddleware // TODO Switch to err no negotiator
	}

	// add current part to context
	// TODO address part is a very bad name, find better one to describe address parts
	// TODO do we even still need context with values?
	mctx := context.WithValue(ctx, ContextKeyAddressPart, ns)
	// and execute them

	if err := ng.Negotiate(mctx, c); err != nil {
		return err
	}

	// and move on
	return f.Next(ctx, c)
}
