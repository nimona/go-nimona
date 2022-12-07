package nimona

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mr-tron/base58"
)

type NodeAddr struct {
	Host      string
	Port      int
	Transport string
	PublicKey []byte
}

func (a *NodeAddr) Parse(addr string) error {
	regex := regexp.MustCompile(`nimona://(?:([\w\d]+@))?([\w\d]+):([\w\d\.]+):(\d+)`)
	matches := regex.FindStringSubmatch(addr)
	if len(matches) != 5 {
		return errors.New("invalid input string")
	}

	publicKey := matches[1]
	transport := matches[2]
	host := matches[3]
	port, err := strconv.Atoi(matches[4])
	if err != nil {
		return err
	}

	var key []byte
	if publicKey != "" {
		publicKey = strings.TrimSuffix(publicKey, "@")
		key, err = base58.Decode(publicKey)
		if err != nil {
			return err
		}
	}

	a.Host = host
	a.Port = port
	a.Transport = transport
	a.PublicKey = key

	return nil
}

func (a NodeAddr) Address() string {
	if a.Host == "" || a.Port == 0 {
		return ""
	}
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

func (a NodeAddr) Network() string {
	return a.Transport
}

func (a NodeAddr) String() string {
	b := strings.Builder{}
	b.WriteString("nimona://")
	if a.PublicKey != nil {
		b.WriteString(base58.Encode(a.PublicKey))
		b.WriteString("@")
	}
	b.WriteString(a.Transport)
	b.WriteString(":")
	b.WriteString(a.Host)
	b.WriteString(":")
	b.WriteString(strconv.Itoa(a.Port))
	return b.String()
}

func NewNodeAddr(transport, host string, port int) NodeAddr {
	return NodeAddr{
		Host:      host,
		Port:      port,
		Transport: transport,
	}
}
