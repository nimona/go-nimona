package net

import (
	"io"

	"nimona.io/pkg/crypto"
)

type Connection struct {
	Conn io.ReadWriteCloser
	RemotePeerKey *crypto.PublicKey
	IsIncoming    bool
}
