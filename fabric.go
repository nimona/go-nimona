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

func New(peerID string) *Fabric {
	return &Fabric{
		middleware: map[string]Middleware{
			"nimona": &IdentityMiddleware{
				Local: peerID,
			},
			"ping": &PingMiddleware{},
		},
		transports: map[string]Transport{},
	}
}

type Fabric struct {
	middleware map[string]Middleware
	transports map[string]Transport
}

func (f *Fabric) DialContext(ctx context.Context, addr string) (Conn, error) {
	// TODO validate the address

	// split address into parts
	aps := strings.Split(addr, "/")

	// first part should be the address we need to dial
	// TODO or we might need to resolve it in case it's something like `peer:xxx`
	tad := aps[0]

	// TODO for now we just care about tcp
	if !strings.HasPrefix(tad, "tcp:") {
		return nil, ErrNoTransport
	}

	// TODO find and use correct transport, or go through all of them
	// dial with transport
	tcon, err := net.Dial("tcp", strings.Replace(tad, "tcp:", "", 1))
	if err != nil {
		fmt.Println("Could not connect to server", err)
		return nil, err
	}

	// wrap the plain net.Conn around our own Conn that adds Get and Set values
	// for the various middlware
	// TODO Move this into the transports maybe?
	conn := wrapConn(tcon)

	// TODO close connection when done if something errored
	// defer go func() { if .. conn.Close() }()

	// TODO address should be split into:
	// * one transport
	// * zero or more middleware
	// * one protocol
	// once we have the list of all these, we need to make sure that all
	// middleware were executed successfully before returning the conn, or error

	// assuming addr looked like `tcp:127.0.0.1:3000/nimona:peer-1/echo`
	// once we are connected to the transport part we need to negotiate and
	// go through the various middleware

	// go through all middleware
	for _, mad := range aps[1:] {
		// find if we have middleware for them
		pms := strings.Split(mad, ":")
		mdn := pms[0]
		mid, ok := f.middleware[mdn]
		if !ok {
			return conn, ErrNoSuchMiddleware
		}

		// ask remote to select this middleware
		if err := f.Select(conn, mdn); err != nil {
			return conn, err
		}

		// and execute them
		mcon, err := mid.Negotiate(ctx, conn, strings.Join(pms[1:], ":"))
		if err != nil {
			return conn, err
		}

		// finally replace our conn with the new returned conn
		conn = mcon
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

	i := 0
	// handle all middleware that we are being given
	for {
		fmt.Println("#", i)
		i++
		// get next protocol name
		prot, err := f.HandleSelect(conn)
		if err != nil {
			return err
		}

		// check if there is a middleware for this
		mid, ok := f.middleware[prot]
		if !ok {
			return ErrNoSuchMiddleware
		}

		// execute middleware
		mcon, err := mid.Handle(context.Background(), conn)
		if err != nil {
			fmt.Println(prot, "FAILED with ", err)
			return err
		}

		conn = mcon
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
