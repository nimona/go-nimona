package nimona

import "fmt"

type NodeAddr struct {
	host      string
	port      int
	transport string
}

func (a *NodeAddr) Address() string {
	if a.host == "" || a.port == 0 {
		return ""
	}
	return fmt.Sprintf("%s:%d", a.host, a.port)
}

func (a *NodeAddr) Network() string {
	return a.transport
}

func (a *NodeAddr) String() string {
	return fmt.Sprintf("%s://%s:%d", a.transport, a.host, a.port)
}

func NewNodeAddr(transport, host string, port int) NodeAddr {
	return NodeAddr{
		host:      host,
		port:      port,
		transport: transport,
	}
}
