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
)

func New() *Fabric {
	return &Fabric{}
}

type Fabric struct {
	middleware []Middleware
	transports []Transport
}

func (f *Fabric) AddTransport(tr Transport) error {
	f.transports = append(f.transports, tr)
	return nil
}

func (f *Fabric) AddMiddleware(md Middleware) error {
	f.middleware = append(f.middleware, md)
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

	// put the stack in the conn
	for _, prt := range stack {
		conn.PushStack(prt)
	}

	// TODO close connection when done if something errored
	// defer go func() { if .. conn.Close() }()

	// once we are connected to the transport part we need to negotiate and
	// go through the various middleware

	// go through all the parts of the address that are left in the stack
	for len(conn.GetStack()) > 0 {
		// get next part from the stack
		prt := conn.GetStack()[0]
		// go through all middleware to find what can negotiate for this part
		hnd := false
		for _, mid := range f.middleware {
			if !mid.CanNegotiate(prt) {
				continue
			}
			fmt.Println("Found mid to neg", mid)
			// ask remote to select this middleware
			// TODO should we be sending the whole part (`smth:param`) or
			// just the type of it (`smth`)?
			if err := f.Select(conn, prt); err != nil {
				return conn, err
			}
			// and execute them
			mcon, err := mid.Negotiate(ctx, conn)
			if err != nil {
				return conn, err
			}
			// finally replace our conn with the new returned conn
			conn = mcon
			// we handled this part of the stack
			hnd = true
			// pop item from stack
			conn.PopStack()
			// and move on
			break
		}
		if !hnd {
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
		for _, mid := range f.middleware {
			if !mid.CanHandle(prot) {
				continue
			}
			// execute middleware
			mcon, err := mid.Handle(context.Background(), conn)
			if err != nil {
				return err
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
