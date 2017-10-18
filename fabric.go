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

	// somewhere to keep our connection
	var conn Conn

	// go through the various tranports
	for _, trn := range f.transports {
		if trn.CanDial(stack[0]) {
			tcon, err := trn.DialContext(ctx, stack[0])
			if err != nil {
				// TODO log and handle error correctly
				fmt.Println("Could not dial", stack[0], err)
				continue
			}
			// if we managed to connect then wrap the plain net.Conn around
			// our own Conn that adds Get and Set values for the various middlware
			// TODO Move this into the transports maybe?
			conn = wrapConn(tcon)
		}
	}

	// TODO this shouldn't fail yet, if transports fail, try all middleware to
	// resolve the address, retry all transports, and then fail
	if conn == nil {
		return nil, errors.New("All transports failed")
	}

	// pop first part from stack since we successfully connected
	stack = stack[1:]

	// TODO close connection when done if something errored
	// defer go func() { if .. conn.Close() }()

	// once we are connected to the transport part we need to negotiate and
	// go through the various middleware

	// go through all the parts of the address that are left in the stack
	// TODO figure out how to deal with the last item in the stack
	// ^ this is a weird one.
	// let's assume the client is dialing `tcp:127.0.0.1:3000/nimona:SERVER/ping`
	// it will connect via TCP transport, will go through the `nimona` middlware,
	// and then the connection should probably be returned to the called so it
	// can use it.
	// on the server's side it will end up on the `ping` handler which is fine.
	// the issue is that if we do `for len(stack) > 1 {` it means
	// that there is no way to have a negotiator be the last part of the stack.
	// maybe it is better to allow the last part of the stack to be processed
	// normally but simply not fail with ErrNoSuchMiddleware if a negotiator
	// doesn't exist.
	// both cases seem wrong though.
	for len(stack) > 0 {
		// get next part from the stack
		prt := stack[0]
		// go through all middleware to find what can negotiate for this part
		hnd := false
		for _, mid := range f.negotiators {
			if !mid.CanNegotiate(prt) {
				continue
			}
			// ask remote to select this middleware
			// TODO should we be sending the whole part (`smth:param`) or
			// just the type of it (`smth`)?
			if err := f.Select(conn, prt); err != nil {
				return conn, err
			}
			// add current part to context
			// TODO address part is a very bad name, find better one to describe address parts
			mctx := context.WithValue(ctx, ContextKeyAddressPart, prt)
			// and execute them
			mcon, err := mid.Negotiate(mctx, conn)
			if err != nil {
				return conn, err
			}
			// finally replace our conn with the new returned conn
			conn = mcon
			// we handled this part of the stack
			hnd = true
			// pop item from stack
			stack = stack[1:]
			// and move on
			break
		}
		if !hnd {
			if len(stack) == 1 {
				fmt.Println("Got last item in stack and have no negotiator, selecting and returning conn", prt)
				// ask remote to select this protocol
				if err := f.Select(conn, prt); err != nil {
					return conn, err
				}
				return conn, nil
			}
			return nil, ErrNoSuchMiddleware
		}
	}

	return conn, nil
}

func (f *Fabric) Select(conn Conn, protocol string) error {
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

	if string(resp) != "ok" {
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
		go f.handleRequest(conn)
	}
	// }()
	// return nil
}

// Handles incoming requests.
func (f *Fabric) handleRequest(tcon net.Conn) error {
	// a client initiated a connection
	fmt.Println("New incoming connection")

	// wrap net.Conn in Conn
	conn := wrapConn(tcon)

	// close the connection when we're done
	defer conn.Close()

	// handle all middleware that we are being given
	for {
		// get next protocol name
		prot, err := f.HandleSelect(conn)
		if err != nil {
			return err
		}

		// check if there is a middleware for this
		hnd := false
		for _, mid := range f.handlers {
			if !mid.CanHandle(prot) {
				continue
			}
			// execute middleware
			mcon, err := mid.Handle(context.Background(), conn)
			if err != nil {
				// TODO should we be closing the connection here?
				conn.Close()
				return err
			}
			// TODO find a way to figure out if the connection was closed
			// if the handler didn't return a connection exit cleanly
			if mcon == nil {
				// but first try to close the connection
				// TODO should we shallow the error?
				// TODO what happens if connection is already closed?
				conn.Close()
				return nil
			}
			// we handled this part of the stack
			hnd = true
			// update connection
			conn = mcon
			// and move on
			break
		}
		if !hnd {
			return ErrNoSuchMiddleware
		}
	}

	return nil
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
	fmt.Println("HandleSelect: Writing ok")

	if err := WriteToken(conn, []byte("ok")); err != nil {
		fmt.Println("Could not write identity response", err)
		return "", err
	}

	return string(prot), nil
}
