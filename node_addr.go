package nimona

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type NodeAddr struct {
	host      string
	port      int
	transport string
	extra     string
}

func (a NodeAddr) Address() string {
	if a.host == "" || a.port == 0 {
		return ""
	}
	return fmt.Sprintf("%s:%d", a.host, a.port)
}

func (a NodeAddr) Network() string {
	return a.transport
}

func (a NodeAddr) String() string {
	return fmt.Sprintf("nimona://%s:%s:%d", a.transport, a.host, a.port)
}

func (a *NodeAddr) Parse(addr string) error {
	if !strings.HasPrefix(addr, "nimona://") {
		return errors.New("unsupported scheme")
	}

	addr = strings.TrimPrefix(addr, "nimona://")

	transport, addr, ok := strings.Cut(addr, ":")
	if !ok {
		return errors.New("invalid address, can't find transport")
	}

	host, addr, ok := strings.Cut(addr, ":")
	if !ok {
		return errors.New("invalid address, can't find host")
	}

	extra := ""
	portStr, extra, _ := strings.Cut(addr, "/")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return errors.New("invalid port")
	}

	a.host = host
	a.port = port
	a.transport = transport
	a.extra = extra

	return nil
}

func NewNodeAddr(transport, host string, port int) NodeAddr {
	return NodeAddr{
		host:      host,
		port:      port,
		transport: transport,
	}
}
