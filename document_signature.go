package nimona

import (
	"fmt"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
)

type Signature struct {
	Signer PeerKey `cborgen:"signer"`
	X      []byte  `cborgen:"x"`
}

func NewDocumentSignature(sk PrivateKey, hash DocumentHash) (*Signature, error) {
	sig, err := sk.Sign(hash[:], &ed25519.Options{})
	if err != nil {
		return nil, fmt.Errorf("error signing document: %w", err)
	}

	return sig, nil
}

func VerifySignature(sig *Signature, hash DocumentHash) error {
	ok := ed25519.Verify(ed25519.PublicKey(sig.Signer.PublicKey), hash[:], sig.X)
	if !ok {
		return fmt.Errorf("error verifying document signature")
	}

	return nil
}
