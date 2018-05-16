package mesh

import (
	"crypto/ecdsa"
	"errors"
	"log"
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
	privateKey, err := ethcrypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	if err := ethcrypto.SaveECDSA(keyPath, privateKey); err != nil {
		return nil, err
	}

	return privateKey, nil
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
