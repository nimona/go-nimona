package crypto

import "nimona.io/go/encoding"

// Sign any block (container) with given key and return a signature block (container)
func Sign(o *encoding.Object, key *Key) error {
	o.SetSignerKey(key.ToObject())

	sig, err := NewSignature(key, AlgorithmObjectHash, o)
	if err != nil {
		return err
	}

	o.SetSignature(sig.ToObject())

	return nil
}
