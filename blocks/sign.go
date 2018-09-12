package blocks

import (
	"errors"

	"nimona.io/go/codec"
	"nimona.io/go/crypto"
)

func Sign(v Typed, key *crypto.Key) (*crypto.Signature, error) {
	p, err := Pack(v, SignWith(key))
	if err != nil {
		return nil, err
	}

	if p.Signature == nil {
		return nil, errors.New("no signature")
	}

	bs, err := blockMapToBlock(p.Signature)
	if err != nil {
		return nil, err
	}

	s, err := Unpack(bs)
	if err != nil {
		return nil, err
	}

	return s.(*crypto.Signature), nil
}

func signPacked(p *Block, key *crypto.Key) (*crypto.Signature, error) {
	digest, err := getDigest(p)
	if err != nil {
		return nil, err
	}

	signature, err := crypto.NewSignature(key, crypto.ES256, digest)
	if err != nil {
		return nil, err
	}

	return signature, nil
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
