package net

import (
	"net"

	"nimona.io/pkg/context"
	"nimona.io/pkg/crypto"
)

type Transport interface {
	Dial(
		ctx context.Context,
		address string,
	) (*Connection, error)
	Listen(
		ctx context.Context,
		bindAddress string,
		key crypto.PrivateKey,
	) (net.Listener, error)
}
