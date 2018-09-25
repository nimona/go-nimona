package primitives

import (
	"nimona.io/go/codec"
)

func Sign(p *Block, key *Key) error {
	digest, err := getDigest(p)
	if err != nil {
		return err
	}

	signature, err := NewSignature(key, ES256, digest)
	if err != nil {
		return err
	}

	p.Signature = signature

	return nil
}

func getDigest(p *Block) ([]byte, error) {
	b := &Block{
		Type:        p.Type,
		Annotations: p.Annotations,
		Payload:     p.Payload,
	}

	digest, err := codec.Marshal(b)
	if err != nil {
		return nil, err
	}

	return digest, nil
}
