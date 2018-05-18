package mesh

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"log"
	"math/big"
	"os"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func LoadOrCreatePrivateKey(keyPath string) (*ecdsa.PrivateKey, error) {
	if keyPath == "" {
		return nil, errors.New("missing key path")
	}

	if _, err := os.Stat(keyPath); err == nil {
		return ethcrypto.LoadECDSA(keyPath)
	}

	log.Printf("* Key path does not exist, creating new key in '%s'\n", keyPath)
	privateKey, err := CreatePrivateKey()
	if err != nil {
		return nil, err
	}

	if err := ethcrypto.SaveECDSA(keyPath, privateKey); err != nil {
		return nil, err
	}

	return privateKey, nil
}

func CreatePrivateKey() (*ecdsa.PrivateKey, error) {
	return ethcrypto.GenerateKey()
}

func DecocdePublicKey(bs []byte) *ecdsa.PublicKey {
	pk := ethcrypto.ToECDSAPub(bs)
	return pk
}

func EncodePublicKey(pk ecdsa.PublicKey) []byte {
	return ethcrypto.FromECDSAPub(&pk)
}

func IDFromPublicKey(pk ecdsa.PublicKey) string {
	return ethcrypto.PubkeyToAddress(pk).String()
}

func Sign(k *ecdsa.PrivateKey, b []byte) ([]byte, error) {
	digest := sha256.Sum256(b)
	r, s, err := ecdsa.Sign(rand.Reader, k, digest[:])
	if err != nil {
		return nil, err
	}

	params := k.Curve.Params()
	curveOrderByteSize := params.P.BitLen() / 8
	rBytes, sBytes := r.Bytes(), s.Bytes()
	signature := make([]byte, curveOrderByteSize*2)
	copy(signature[curveOrderByteSize-len(rBytes):], rBytes)
	copy(signature[curveOrderByteSize*2-len(sBytes):], sBytes)
	return signature, nil
}

func Verify(pk *ecdsa.PublicKey, b []byte, sn []byte) (bool, error) {
	digest := sha256.Sum256(b)
	curveOrderByteSize := pk.Curve.Params().P.BitLen() / 8
	r, s := new(big.Int), new(big.Int)
	r.SetBytes(sn[:curveOrderByteSize])
	s.SetBytes(sn[curveOrderByteSize:])
	return ecdsa.Verify(pk, digest[:], r, s), nil
}
