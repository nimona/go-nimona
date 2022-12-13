package nimona

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/mr-tron/base58"
	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
)

const (
	PeerAddressPrefix = "nimona://peer:addr:"
	PeerHandlePrefix  = "nimona://peer:handle:"
	PeerKeyPrefix     = "nimona://peer:key:"
)

type NodeAddr struct {
	address   string
	network   string
	publicKey []byte
}

func (a *NodeAddr) Parse(addr string) error {
	regex := regexp.MustCompile(
		PeerAddressPrefix + `(?:([\w\d]+@))?([\w\d]+):([\w\d\.]+):(\d+)`,
	)
	matches := regex.FindStringSubmatch(addr)
	if len(matches) != 5 {
		return errors.New("invalid input string")
	}

	publicKey := matches[1]
	transport := matches[2]
	host := matches[3]
	port := matches[4]

	if publicKey != "" {
		publicKey = strings.TrimSuffix(publicKey, "@")
		key, err := base58.Decode(publicKey)
		if err != nil {
			return err
		}
		a.publicKey = key
	}

	a.network = transport
	a.address = fmt.Sprintf("%s:%s", host, port)

	return nil
}

func (a NodeAddr) Network() string {
	return a.network
}

func (a NodeAddr) String() string {
	b := strings.Builder{}
	b.WriteString(PeerAddressPrefix)
	if a.publicKey != nil {
		b.WriteString(base58.Encode(a.publicKey))
		b.WriteString("@")
	}
	b.WriteString(a.network)
	b.WriteString(":")
	b.WriteString(a.address)
	return b.String()
}

func (a NodeAddr) Address() string {
	return a.address
}

func (a NodeAddr) PublicKey() ed25519.PublicKey {
	return a.publicKey
}

func NewNodeAddr(transport, address string) NodeAddr {
	return NodeAddr{
		address: address,
		network: transport,
	}
}

func NewNodeAddrWithKey(transport, address string, key []byte) NodeAddr {
	return NodeAddr{
		address:   address,
		network:   transport,
		publicKey: key,
	}
}
