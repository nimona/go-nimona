package nimona

import (
	"fmt"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
)

type Signature struct {
	Signer PeerKey `nimona:"signer"`
	X      []byte  `nimona:"x"`
}

func NewDocumentSignature(sk PrivateKey, hash DocumentHash) *Signature {
	sig, err := sk.Sign(hash[:], &ed25519.Options{})
	if err != nil {
		panic(fmt.Errorf("error signing document: %w", err))
	}

	return sig
}

func VerifySignature(sig *Signature, hash DocumentHash) error {
	ok := ed25519.Verify(ed25519.PublicKey(sig.Signer.PublicKey), hash[:], sig.X)
	if !ok {
		return fmt.Errorf("error verifying document signature")
	}

	return nil
}
