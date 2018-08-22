package blocks

func Sign(block *Block, key Key) error {
	digest, err := MarshalClean(block)
	if err != nil {
		return err
	}

	signature, err := NewSignature(key, ES256, digest)
	if err != nil {
		return err
	}

	block.Signature = signature
	return nil
}
