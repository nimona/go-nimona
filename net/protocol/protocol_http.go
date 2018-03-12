package protocol

import (
	"context"
	"net"
	"net/http"
)

func NewHTTPClient(c net.Conn) (*http.Client, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return c, nil
			},
		},
	}
	return client, nil
}

func NewHTTPServer(c net.Conn, hn http.Handler) error {
	server := &http.Server{
		Addr:    ":http",
		Handler: hn,
	}
	listener := &httpConnListener{
		conn: c,
	}
	return server.Serve(listener)
}

type httpConnListener struct {
	conn net.Conn
}

func (s *httpConnListener) Accept() (net.Conn, error) {
	// TODO Check conn
	return s.conn, nil
}

func (s *httpConnListener) Close() error {
	// TODO Close conn
	return nil
}

func (s *httpConnListener) Addr() net.Addr {
	return s.conn.LocalAddr()
}
