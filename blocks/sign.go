package blocks

import (
	"nimona.io/go/codec"
	"nimona.io/go/crypto"
)

func Signature(v Typed, key *crypto.Key) (*crypto.Signature, error) {
	p, err := Pack(v, SignWith(key))
	if err != nil {
		return nil, err
	}

	ps := p.Signature
	s, err := UnpackDecodeBase58(ps)
	if err != nil {
		return nil, err
	}

	return s.(*crypto.Signature), nil
}

func signPacked(p *Block, key *crypto.Key) (string, error) {
	digest, err := getDigest(p)
	if err != nil {
		return "", err
	}

	signature, err := crypto.NewSignature(key, crypto.ES256, digest)
	if err != nil {
		return "", err
	}

	bs, err := PackEncodeBase58(signature)
	if err != nil {
		return "", err
	}

	return bs, nil
}

func getDigest(p *Block) ([]byte, error) {
	b := &Block{
		Type:    p.Type,
		Payload: p.Payload,
	}

	digest, err := codec.Marshal(b)
	if err != nil {
		return nil, err
	}

	return digest, nil
}
