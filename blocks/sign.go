package blocks

// TODO should sign return?
func Sign(block *Block, key *Key) error {
	digest, err := getDigest(block)
	if err != nil {
		return err
	}

	signature, err := NewSignature(key, ES256, digest)
	if err != nil {
		return err
	}

	bs, err := Marshal(signature)
	if err != nil {
		return err
	}

	block.Signature = bs
	return nil
}

func getDigest(block *Block) ([]byte, error) {
	b := &Block{
		Type:     block.Type,
		Metadata: block.Metadata,
		Payload:  block.Payload,
	}

	digest, err := MarshalBlock(b)
	if err != nil {
		return nil, err
	}

	return digest, nil
}
