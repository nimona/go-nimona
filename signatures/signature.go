package signatures

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"github.com/nimona/go-nimona/blocks"
	"github.com/nimona/go-nimona/keys"
)

var (
	// ErrInvalidBlockType is returned when the signature being verified
	// is not an encoded block of type "signature".
	ErrInvalidBlockType = errors.New("invalid block type")
	// ErrAlgorithNotImplemented is returned when the algorithm specified
	// has not been implemented
	ErrAlgorithNotImplemented = errors.New("algorithm not implemented")
	// ErrCouldNotVerify is returned when the signature doesn't matches the
	// given key
	ErrCouldNotVerify = errors.New("could not verify signature")
)

// Algorithm for Signature
type Algorithm string

// Supported values for Algorithm
const (
	ES256          Algorithm = "ES256" // ECDSA using P-256 and SHA-256
	ES384          Algorithm = "ES384" // ECDSA using P-384 and SHA-384
	ES512          Algorithm = "ES512" // ECDSA using P-521 and SHA-512
	HS256          Algorithm = "HS256" // HMAC using SHA-256
	HS384          Algorithm = "HS384" // HMAC using SHA-384
	HS512          Algorithm = "HS512" // HMAC using SHA-512
	InvalidKeyType Algorithm = ""      // Invalid KeyType
	NoSignature    Algorithm = "none"  // No signature
	PS256          Algorithm = "PS256" // RSASSA-PSS using SHA256 and MGF1-SHA256
	PS384          Algorithm = "PS384" // RSASSA-PSS using SHA384 and MGF1-SHA384
	PS512          Algorithm = "PS512" // RSASSA-PSS using SHA512 and MGF1-SHA512
	RS256          Algorithm = "RS256" // RSASSA-PKCS-v1.5 using SHA-256
	RS384          Algorithm = "RS384" // RSASSA-PKCS-v1.5 using SHA-384
	RS512          Algorithm = "RS512" // RSASSA-PKCS-v1.5 using SHA-512
)

// String returns the string representation of a Algorithm
func (v Algorithm) String() string {
	return string(v)
}

// Sign returns a signature given some bytes and a private key
func Sign(key keys.Key, alg Algorithm, data []byte) ([]byte, error) {
	mKey := key.Materialize()
	pKey, ok := mKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("only ecdsa keys are currently supported")
	}

	if alg != ES256 {
		return nil, ErrAlgorithNotImplemented
	}

	// TODO implement more algorithms
	digest := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, pKey, digest[:])
	if err != nil {
		return nil, err
	}

	params := pKey.Curve.Params()
	curveOrderByteSize := params.P.BitLen() / 8
	rBytes, sBytes := r.Bytes(), s.Bytes()
	signature := make([]byte, curveOrderByteSize*2)
	copy(signature[curveOrderByteSize-len(rBytes):], rBytes)
	copy(signature[curveOrderByteSize*2-len(sBytes):], sBytes)

	block := blocks.NewBlock("signature", Headers{
		Type:      alg.String(),
		Signature: signature,
	})

	return blocks.Marshal(block)
}

// Verify the signature of some data given a public
func Verify(key keys.Key, data, signatureBlock []byte) error {
	mKey := key.Materialize()
	pKey, ok := mKey.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("only ecdsa public keys are currently supported")
	}

	block, err := blocks.Unmarshal(signatureBlock)
	if err != nil {
		return err
	}

	if block.Type != "signature" {
		return ErrInvalidBlockType
	}

	// TODO implement more algorithms
	if block.Payload.(Headers).Type != ES256.String() {
		return ErrAlgorithNotImplemented
	}

	headers := block.Payload.(Headers)
	signature := headers.Signature

	digest := sha256.Sum256(data)
	rBytes := new(big.Int).SetBytes(signature[0:32])
	sBytes := new(big.Int).SetBytes(signature[32:64])
	fmt.Printf("___ %x %x %x", signature, rBytes, sBytes)
	if ok := ecdsa.Verify(pKey, digest[:], rBytes, sBytes); !ok {
		return ErrCouldNotVerify
	}

	return nil
}
