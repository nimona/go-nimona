package net

import (
	"crypto/rsa"
	"crypto/x509"
)

type ID []byte

var (
	identityIDPrefix = []byte("00x")
	peerIDPrefix     = []byte("01x")
)

func (id ID) GetPublicKey() *rsa.PublicKey {
	publicKey, _ := x509.ParsePKCS1PublicKey(id)
	return publicKey
}
