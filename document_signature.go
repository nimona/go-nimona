package nimona

import (
	"fmt"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
)

type Signature struct {
	Signer PeerID `cborgen:"signer"`
	X      []byte `cborgen:"x"`
}

func NewDocumentSignature(sk ed25519.PrivateKey, hash Hash) (*Signature, error) {
	sig, err := sk.Sign(nil, hash[:], &ed25519.Options{})
	if err != nil {
		return nil, fmt.Errorf("error signing document: %w", err)
	}

	return &Signature{
		Signer: NewPeerID(sk.Public().(ed25519.PublicKey)),
		X:      sig[:],
	}, nil
}

func VerifyDocument(cborBytes []byte, sig *Signature) error {
	h, err := NewDocumentHashFromCBOR(cborBytes)
	if err != nil {
		return fmt.Errorf("error creating document hash: %w", err)
	}

	ok := ed25519.Verify(sig.Signer.PublicKey, h[:], sig.X[:])
	if !ok {
		return fmt.Errorf("error verifying document signature: %w", err)
	}

	return nil
}
