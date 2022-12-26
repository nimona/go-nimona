package nimona

import (
	"github.com/mr-tron/base58"
	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
)

func PublicKeyToBase58(pk ed25519.PublicKey) string {
	return base58.Encode(pk)
}

func PublicKeyFromBase58(pk string) (ed25519.PublicKey, error) {
	return base58.Decode(pk)
}
