package main

import (
	"context"
	"fmt"

	fabric "github.com/nimona/go-nimona-fabric"
)

func main() {
	f := fabric.New()
	f.AddTransport(&fabric.TCP{})
	f.AddMiddleware(&fabric.IdentityMiddleware{Local: "CLIENT"})

	// make a new connection to the the server's ping handler
	conn, err := f.DialContext(context.Background(), "tcp:127.0.0.1:3000/nimona:SERVER/ping")
	if err != nil {
		fmt.Println("Dial error", err)
		return
	}

	// close conection when done
	defer conn.Close()

	// send ping
	fmt.Println("Ping: Writing ping...")
	if err := fabric.WriteToken(conn, []byte("PING")); err != nil {
		fmt.Println("Could not ping", err)
		return
	}

	fmt.Println("Ping: Wrote ping")

	// get pong
	fmt.Println("Ping: Reading pong...")
	pong, err := fabric.ReadToken(conn)
	if err != nil {
		fmt.Println("Could not read remote pong", err)
		return
	}

	fmt.Println("Ping: Read pong:", string(pong))
}
