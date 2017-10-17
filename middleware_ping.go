package fabric

import (
	"context"
	"fmt"
)

const (
	PingKey = "ping"
)

type PingMiddleware struct{}

func (m *PingMiddleware) Handle(ctx context.Context, conn Conn) (Conn, error) {
	// client pings
	fmt.Println("Ping.Handle: Reading ping")
	ping, err := ReadToken(conn)
	if err != nil {
		fmt.Println("Could not read remote ping", err)
		return nil, err
	}

	fmt.Println("Ping.Handle: Read ping:", string(ping))

	// we pong back
	fmt.Println("Ping.Handle: Writing pong")
	if err := WriteToken(conn, []byte("pong")); err != nil {
		fmt.Println("Could not pong", err)
		return nil, err
	}

	// return connection as it was
	return conn, nil
}

func (m *PingMiddleware) CanHandle(addr string) bool {
	parts := addrSplit(addr)
	return parts[0][0] == PingKey
}

func (m *PingMiddleware) Negotiate(ctx context.Context, conn Conn) (Conn, error) {
	// we ping
	fmt.Println("Ping.Negotiate: Writing ping")
	if err := WriteToken(conn, []byte("ping")); err != nil {
		fmt.Println("Could not ping", err)
		return nil, err
	}

	// remote pongs
	fmt.Println("Ping.Negotiate: Reading pong")
	pong, err := ReadToken(conn)
	if err != nil {
		fmt.Println("Could not read remote pong", err)
		return nil, err
	}

	fmt.Println("Ping.Negotiate: Read pong:", string(pong))

	fmt.Println("Ping.Negotiate: Closing connection")
	conn.Close()

	return conn, nil
}

func (m *PingMiddleware) CanNegotiate(addr string) bool {
	parts := addrSplit(addr)
	return parts[0][0] == PingKey
}
