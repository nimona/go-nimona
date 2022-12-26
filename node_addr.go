package nimona

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type NodeAddr struct {
	_         string `cborgen:"$type,const=core/node.address"`
	Address   string `cborgen:"address"`
	Network   string `cborgen:"network"`
	PublicKey []byte `cborgen:"publicKey"`
}

func ParseNodeAddr(addr string) (*NodeAddr, error) {
	regex := regexp.MustCompile(
		ResourceTypePeerAddress.String() +
			`(?:([\w\d]+@))?([\w\d]+):([\w\d\.]+):(\d+)`,
	)
	matches := regex.FindStringSubmatch(addr)
	if len(matches) != 5 {
		return nil, errors.New("invalid input string")
	}

	publicKey := matches[1]
	transport := matches[2]
	host := matches[3]
	port := matches[4]

	a := &NodeAddr{}
	if publicKey != "" {
		publicKey = strings.TrimSuffix(publicKey, "@")
		key, err := PublicKeyFromBase58(publicKey)
		if err != nil {
			return nil, fmt.Errorf("invalid public key, %w", err)
		}
		a.PublicKey = key
	}

	a.Network = transport
	a.Address = fmt.Sprintf("%s:%s", host, port)

	return a, nil
}

func (a NodeAddr) String() string {
	b := strings.Builder{}
	b.WriteString(ResourceTypePeerAddress.String())
	if a.PublicKey != nil {
		b.WriteString(PublicKeyToBase58(a.PublicKey))
		b.WriteString("@")
	}
	b.WriteString(a.Network)
	b.WriteString(":")
	b.WriteString(a.Address)
	return b.String()
}
