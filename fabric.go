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
	return &Fabric{}
}

type Fabric struct {
	handlers    []Handler
	negotiators []Negotiator
	transports  []Transport
}

func (f *Fabric) AddTransport(tr Transport) error {
	f.transports = append(f.transports, tr)
	return nil
}

func (f *Fabric) AddMiddleware(md Middleware) error {
	if err := f.AddHandler(md); err != nil {
		return err
	}
	return f.AddNegotiator(md)
}

func (f *Fabric) AddHandler(hn Handler) error {
	f.handlers = append(f.handlers, hn)
	return nil
}

func (f *Fabric) AddHandlerFunc(protocol string, hf HandlerFunc) error {
	hn := &simpleHandler{
		protocol: protocol,
		handler:  hf,
	}
	f.handlers = append(f.handlers, hn)
	return nil
}

func (f *Fabric) AddNegotiator(ng Negotiator) error {
	f.negotiators = append(f.negotiators, ng)
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
	if err := f.Next(ctx, conn); err != nil {
		fmt.Println("Error on conn.Next()", err)
		return nil, err
	}

	return conn, nil
}

func (f *Fabric) Select(conn net.Conn, protocol string) error {
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

	// handle all middleware that we are being given
	for {
		fmt.Println("handleRequest: Waiting for next protocol")
		// get next protocol name
		prot, err := f.HandleSelect(c)
		if err != nil {
			return err
		}

		fmt.Println("handleRequest: Got next protocol", prot)

		// check if there is a middleware for this
		hnd := false
		for _, mid := range f.handlers {
			if !mid.CanHandle(prot) {
				continue
			}
			// execute middleware
			if err := mid.Handle(context.Background(), c); err != nil {
				// TODO should we be closing the connection here?
				c.Close()
				return err
			}
			// mark as handled
			fmt.Println("handleRequest: Handled protocol", prot)
			hnd = true
			// and move on
			break
		}
		if !hnd {
			fmt.Println("handleRequest: Could not handle", prot)
			return ErrNoSuchMiddleware
		}
	}

	return nil
}

func (f *Fabric) Next(ctx context.Context, c *conn) error {
	// get next protocol
	ns := c.popStack()
	fmt.Println("Processing", ns)

	// are we dont yet?
	if ns == "" {
		return nil
	}

	// go through the various tranports
	for _, trn := range f.transports {
		if trn.CanDial(ns) {
			tcon, err := trn.DialContext(ctx, ns)
			if err != nil {
				// TODO log and handle error correctly
				fmt.Println("Could not dial", ns, err)
				continue
			}
			// if we managed to connect upgrade our connection
			if err := c.Upgrade(tcon); err != nil {
				return err
			}
			// move on
			return f.Next(ctx, c)
		}
	}

	// TODO this shouldn't fail yet, if transports fail, try all middleware to
	// resolve the address, retry all transports, and then fail
	// if conn == nil {
	// 	return nil, errors.New("All transports failed")
	// }

	lc, err := c.GetRawConn()
	if lc == nil || err != nil {
		return errors.New("All transports failed")
	}

	// go through all middleware to find what can negotiate for this part
	for _, mid := range f.negotiators {
		if !mid.CanNegotiate(ns) {
			continue
		}
		// ask remote to select this middleware
		// TODO should we be sending the whole part (`smth:param`) or
		// just the type of it (`smth`)?
		fmt.Println("Selecting", ns)
		if err := f.Select(lc, ns); err != nil {
			return err
		}
		// add current part to context
		// TODO address part is a very bad name, find better one to describe address parts
		mctx := context.WithValue(ctx, ContextKeyAddressPart, ns)
		// and execute them
		if err := mid.Negotiate(mctx, c); err != nil {
			return err
		}
		// and move on
		return f.Next(ctx, c)
	}

	// ask remote to select this protocol
	// TODO should we check if this is the last part of the addr?
	// TODO what hapens if it's not?
	fmt.Println("Got no negotiator, selecting", ns)
	return f.Select(c, ns)
}

func (f *Fabric) HandleSelect(conn Conn) (string, error) {
	// we need to negotiate what they need from us
	fmt.Println("HandleSelect: Reading protocol token")
	// read the next token, which is the request for the next middleware
	prot, err := ReadToken(conn)
	if err != nil {
		fmt.Println("Could not read token", err)
		return "", err
	}

	fmt.Println("HandleSelect: Read protocol token:", string(prot))
	fmt.Println("HandleSelect: Writing protocol as ack")

	if err := WriteToken(conn, prot); err != nil {
		fmt.Println("Could not write protocol ack", err)
		return "", err
	}

	return string(prot), nil
}
