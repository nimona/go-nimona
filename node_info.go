package nimona

import "github.com/oasisprotocol/curve25519-voi/primitives/ed25519"

type NodeInfo struct {
	Addr      NodeAddr
	PublicKey ed25519.PublicKey
}
