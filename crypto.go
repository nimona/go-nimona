package nimona

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/mr-tron/base58"
	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/oasisprotocol/curve25519-voi/primitives/x25519"
)

type (
	KeyPair struct {
		PrivateKey PrivateKey
		PublicKey  PublicKey
	}
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
)

// TODO(geoah) deprecate
func GenerateKey() (PublicKey, PrivateKey, error) {
	pk, sk, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error generating key pair: %w", err)
	}
	return PublicKey(pk), PrivateKey(sk), nil
}

func GenerateKeyPair() (*KeyPair, error) {
	pk, sk, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, fmt.Errorf("error generating key pair: %w", err)
	}
	return &KeyPair{
		PrivateKey: PrivateKey(sk),
		PublicKey:  PublicKey(pk),
	}, nil
}

func (pk PublicKey) String() string {
	return base58.Encode(pk)
}

func (pk PublicKey) IsZero() bool {
	return len(pk) == 0
}

func (pk PublicKey) Equal(other PublicKey) bool {
	return ed25519.PublicKey(pk).Equal(ed25519.PublicKey(other))
}

func (pk PublicKey) X25519() ([]byte, error) {
	px, ok := x25519.EdPublicKeyToX25519(ed25519.PublicKey(pk))
	if !ok {
		return nil, fmt.Errorf("error converting public key to x25519")
	}
	return px, nil
}

func ParsePublicKey(pk string) (PublicKey, error) {
	return base58.Decode(pk)
}

func ParsePublicKeyFromHex(pk string) (PublicKey, error) {
	return hex.DecodeString(pk)
}

func ParsePublicKeyFromBase64(pk string) (PublicKey, error) {
	return base64.StdEncoding.DecodeString(pk)
}

func (sk PrivateKey) Equal(other PrivateKey) bool {
	return ed25519.PrivateKey(sk).Equal(ed25519.PrivateKey(other))
}

func (sk PrivateKey) Sign(message []byte, opts *ed25519.Options) (*Signature, error) {
	sig, err := ed25519.PrivateKey(sk).Sign(nil, message, &ed25519.Options{})
	if err != nil {
		return nil, fmt.Errorf("error signing document: %w", err)
	}

	return &Signature{
		Signer: NewPeerKey(sk.Public()),
		X:      sig,
	}, nil
}

func (sk PrivateKey) Public() PublicKey {
	return PublicKey(ed25519.PrivateKey(sk).Public().(ed25519.PublicKey))
}

func (sk PrivateKey) IsZero() bool {
	return len(sk) == 0
}

func (sk PrivateKey) X25519() ([]byte, error) {
	sx := x25519.EdPrivateKeyToX25519(ed25519.PrivateKey(sk))
	return sx, nil
}

func ParsePrivateKey(sk string) (PrivateKey, error) {
	return base58.Decode(sk)
}

func ParsePrivateKeyFromHex(sk string) (PrivateKey, error) {
	return hex.DecodeString(sk)
}

func ParsePrivateKeyFromBase64(sk string) (PrivateKey, error) {
	return base64.StdEncoding.DecodeString(sk)
}

func SignDocument(rctx *RequestContext, doc *Document) {
	switch {
	case rctx == nil:
		return
	case doc == nil:
		return
	case rctx.PrivateKey.IsZero():
		return
	}

	if doc.Metadata.Owner == nil {
		doc.Metadata.Owner = rctx.Identity
	}

	doc.Metadata.Signature = NewDocumentSignature(
		rctx.PrivateKey,
		NewDocumentHash(doc),
	)
}
