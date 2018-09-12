package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"errors"

	"nimona.io/go/base58"
	"nimona.io/go/codec"
)

var (
	// ErrInvalidBlockType is returned when the signature being verified
	// is not an encoded block of type "signature".
	ErrInvalidBlockType = errors.New("invalid block type")
	// ErrAlgorithNotImplemented is returned when the algorithm specified
	// has not been implemented
	ErrAlgorithNotImplemented = errors.New("algorithm not implemented")
)

// Algorithm for Signature
type Algorithm string

// Supported values for Algorithm
const (
	ES256            Algorithm = "ES256" // ECDSA using P-256 and SHA-256
	ES384            Algorithm = "ES384" // ECDSA using P-384 and SHA-384
	ES512            Algorithm = "ES512" // ECDSA using P-521 and SHA-512
	HS256            Algorithm = "HS256" // HMAC using SHA-256
	HS384            Algorithm = "HS384" // HMAC using SHA-384
	HS512            Algorithm = "HS512" // HMAC using SHA-512
	InvalidAlgorithm Algorithm = ""      // Invalid Algorithm
	NoSignature      Algorithm = "none"  // No signature
	PS256            Algorithm = "PS256" // RSASSA-PSS using SHA256 and MGF1-SHA256
	PS384            Algorithm = "PS384" // RSASSA-PSS using SHA384 and MGF1-SHA384
	PS512            Algorithm = "PS512" // RSASSA-PSS using SHA512 and MGF1-SHA512
	RS256            Algorithm = "RS256" // RSASSA-PKCS-v1.5 using SHA-256
	RS384            Algorithm = "RS384" // RSASSA-PKCS-v1.5 using SHA-384
	RS512            Algorithm = "RS512" // RSASSA-PKCS-v1.5 using SHA-512
)

// String returns the string representation of a Algorithm
func (v Algorithm) String() string {
	return string(v)
}

type Signature struct {
	Key       *Key      `json:"key"`
	Alg       Algorithm `json:"alg"`
	Signature []byte    `json:"sig"`
}

func (s *Signature) GetType() string {
	return "signature"
}

func (s *Signature) GetSignature() *Signature {
	// no signature
	// TODO this is a bit ironic :P
	return nil
}

func (s *Signature) SetSignature(*Signature) {
	// no signature
}

func (s *Signature) GetAnnotations() map[string]interface{} {
	// no annotations
	return map[string]interface{}{}
}

func (s *Signature) SetAnnotations(a map[string]interface{}) {
	// no annotations
}

func (s *Signature) MarshalBlock() (string, error) {
	key := ""
	if s.Key != nil {
		k, err := s.Key.MarshalBlock()
		if err != nil {
			return "", err
		}
		key = k
	}

	block := map[string]interface{}{
		"type": "signature",
		"payload": map[string]interface{}{
			"key": key,
			"alg": s.Alg,
			"sig": s.Signature,
		},
	}

	bytes, err := codec.Marshal(block)
	if err != nil {
		return "", err
	}

	return base58.Encode(bytes), nil
}

func (s *Signature) UnmarshalBlock(b58bytes string) error {
	bytes, err := base58.Decode(b58bytes)
	if err != nil {
		return err
	}

	block := map[string]interface{}{}
	if err := codec.Unmarshal(bytes, &block); err != nil {
		return err
	}

	// HACK too many hacks, make a decoder
	payload := block["payload"].(map[string]interface{})
	if ab, ok := payload["alg"]; ok && ab != nil {
		s.Alg = Algorithm(ab.(string))
	}
	if sb, ok := payload["sig"]; ok && sb != nil {
		s.Signature = sb.([]byte)
	}

	key, ok := payload["key"].(string)
	if ok && key != "" {
		s.Key = &Key{}
		if err := s.Key.UnmarshalBlock(key); err != nil {
			return err
		}
	}

	return nil
}

// NewSignature returns a signature given some bytes and a private key
func NewSignature(key *Key, alg Algorithm, digest []byte) (*Signature, error) {
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
