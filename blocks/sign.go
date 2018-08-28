package blocks

import (
	"fmt"
)

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

	fmt.Println("SIGNED IT WITH", Base58Encode(block.Signature))

	return nil
}

func getDigest(block *Block) ([]byte, error) {
	b := &Block{
		Type:     block.Type,
		Metadata: block.Metadata,
		Payload:  block.Payload,
	}

	// x, _ := json.MarshalIndent(b, "", "  ")
	// fmt.Println(string(x))

	digest, err := MarshalBlock(b)
	if err != nil {
		return nil, err
	}

	return digest, nil
}
