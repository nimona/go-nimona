package connmanager

import (
	"net"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
)

type Transport interface {
	Dial(
		ctx context.Context,
		address string,
	) (*connection, error)
	Listen(
		ctx context.Context,
		bindAddress string,
		key crypto.PrivateKey,
	) (net.Listener, error)
}
