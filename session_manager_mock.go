package nimona

import (
	"io"
	"net"
	"time"
)

type MockConn struct {
	Server net.Conn
	Client net.Conn
}

func NewMockConn() *MockConn {
	sr, cw := io.Pipe()
	cr, sw := io.Pipe()

	return &MockConn{
		Server: &MockConnEndpoint{
			Reader: sr,
			Writer: sw,
		},
		Client: &MockConnEndpoint{
			Reader: cr,
			Writer: cw,
		},
	}
}

type MockAddr struct {
	network string
	address string
}

func (m *MockAddr) Network() string {
	return m.network
}

func (m *MockAddr) String() string {
	return m.address
}

type MockConnEndpoint struct {
	Reader io.Reader
	Writer io.Writer
}

func (m *MockConnEndpoint) Read(b []byte) (n int, err error) {
	return m.Reader.Read(b)
}

func (m *MockConnEndpoint) Write(b []byte) (n int, err error) {
	return m.Writer.Write(b)
}

func (m *MockConnEndpoint) Close() error {
	return nil
}

func (m *MockConnEndpoint) LocalAddr() net.Addr {
	return &MockAddr{
		network: "mock",
		address: "local",
	}
}

func (m *MockConnEndpoint) RemoteAddr() net.Addr {
	return &MockAddr{
		network: "mock",
		address: "remote",
	}
}

func (m *MockConnEndpoint) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockConnEndpoint) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockConnEndpoint) SetWriteDeadline(t time.Time) error {
	return nil
}