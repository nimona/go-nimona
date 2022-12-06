package nimona

import (
	"net"
	"testing"
)

func TestSession(t *testing.T) {
	// Create a listener on localhost
	ln, err := net.Listen("tcp", "localhost:9000")
	if err != nil {
		t.Error(err)
		return
	}
	defer ln.Close()

	// Start the server in a goroutine
	go func() {
		// Accept incoming connections
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()

		// Create a new session for the connection
		session, err := NewSession(conn)
		if err != nil {
			t.Error(err)
			return
		}

		// Read a packet from the connection
		data, err := session.Read()
		if err != nil {
			t.Error(err)
			return
		}
		if string(data) != "Hello, server!" {
			t.Error("Incorrect data received")
			return
		}

		// Write a packet to the connection
		_, err = session.Write([]byte("Hello, client!"))
		if err != nil {
			t.Error(err)
			return
		}
	}()

	// Connect to the server as a client
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Close()

	// Create a new session for the connection
	session, err := NewSession(conn)
	if err != nil {
		t.Error(err)
		return
	}

	// Write a packet to the connection
	_, err = session.Write([]byte("Hello, server!"))
	if err != nil {
		t.Error(err)
		return
	}

	// Read a packet from the connection
	data, err := session.Read()
	if err != nil {
		t.Error(err)
		return
	}
	if string(data) != "Hello, client!" {
		t.Error("Incorrect data received")
		return
	}
}
