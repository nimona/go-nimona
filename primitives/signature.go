package primitives

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"errors"
)

var (
	// ErrInvalidBlockType is returned when the signature being verified
	// is not an encoded block of type "signature".
	ErrInvalidBlockType = errors.New("invalid block type")
	// ErrAlgorithNotImplemented is returned when the algorithm specified
	// has not been implemented
	ErrAlgorithNotImplemented = errors.New("algorithm not implemented")
)

// Supported values for Algorithm
const (
	ES256            = "ES256" // ECDSA using P-256 and SHA-256
	ES384            = "ES384" // ECDSA using P-384 and SHA-384
	ES512            = "ES512" // ECDSA using P-521 and SHA-512
	HS256            = "HS256" // HMAC using SHA-256
	HS384            = "HS384" // HMAC using SHA-384
	HS512            = "HS512" // HMAC using SHA-512
	InvalidAlgorithm = ""      // Invalid Algorithm
	NoSignature      = "none"  // No signature
	PS256            = "PS256" // RSASSA-PSS using SHA256 and MGF1-SHA256
	PS384            = "PS384" // RSASSA-PSS using SHA384 and MGF1-SHA384
	PS512            = "PS512" // RSASSA-PSS using SHA512 and MGF1-SHA512
	RS256            = "RS256" // RSASSA-PKCS-v1.5 using SHA-256
	RS384            = "RS384" // RSASSA-PKCS-v1.5 using SHA-384
	RS512            = "RS512" // RSASSA-PKCS-v1.5 using SHA-512
)

type Signature struct {
	Key       *Key   `json:"key"`
	Alg       string `json:"alg"`
	Signature []byte `json:"sig"`
}

func (s *Signature) Block() *Block {
	return &Block{
		Type: "nimona.io/signature",
		Payload: map[string]interface{}{
			"key": s.Key.Block(),
			"alg": s.Alg,
			"sig": s.Signature,
		},
	}
}

func (s *Signature) FromBlock(block *Block) {
	s.Signature = block.Payload["sig"].([]byte)
	s.Alg = block.Payload["alg"].(string)
	s.Key = &Key{}
	s.Key.FromBlock(BlockFromMap(block.Payload["key"].(map[string]interface{})))
}

// NewSignature returns a signature given some bytes and a private key
func NewSignature(key *Key, alg string, digest []byte) (*Signature, error) {
	if key == nil {
		return nil, errors.New("missing key")
	}

	mKey := key.Materialize()
	if mKey == nil {
		return nil, errors.New("could not materialize")
	}

	pKey, ok := mKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("only ecdsa private keys are currently supported")
	}

	if alg != ES256 {
		return nil, ErrAlgorithNotImplemented
	}

	// TODO implement more algorithms
	hash := sha256.Sum256(digest)
	r, s, err := ecdsa.Sign(rand.Reader, pKey, hash[:])
	if err != nil {
		return nil, err
	}

	params := pKey.Curve.Params()
	curveOrderByteSize := params.P.BitLen() / 8
	rBytes, sBytes := r.Bytes(), s.Bytes()
	signature := make([]byte, curveOrderByteSize*2)
	copy(signature[curveOrderByteSize-len(rBytes):], rBytes)
	copy(signature[curveOrderByteSize*2-len(sBytes):], sBytes)

	return &Signature{
		Key:       key.GetPublicKey(),
		Alg:       alg,
		Signature: signature,
	}, nil
}
