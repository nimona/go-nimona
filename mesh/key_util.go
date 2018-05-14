package mesh

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// Url-safe base64 encode that strips padding
func base64URLEncode(data []byte) string {
	var result = base64.URLEncoding.EncodeToString(data)
	return strings.TrimRight(result, "=")
}

func LoadOrCreatePrivateKey(keyPath string) (*ecdsa.PrivateKey, error) {
	var privateKey *ecdsa.PrivateKey

	// Check if keyPath is empty
	if keyPath == "" {
		keyPath = ".key.pem"
		log.Printf("* Using key from '%s'\n", keyPath)
	}

	// Check if the keyPath exists
	keyExists := true
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		keyExists = false
	}

	// If it exists, try to use it
	if keyExists {
		// load the private key
		var errLoadingKey error
		privateKey, errLoadingKey = loadPrivateKey(keyPath)
		if errLoadingKey != nil {
			return nil, errLoadingKey
		}
	} else {
		log.Printf("* Key path does not exist, creating a key and storing it in '%s'\n", keyPath)
		// generate key pair if we have not been given one
		var errGeneratingPrivateKey error
		privateKey, errGeneratingPrivateKey = generatePrivateKey()
		if errGeneratingPrivateKey != nil {
			return nil, errGeneratingPrivateKey
		}
	}

	if keyExists == false {
		storePrivateKey(privateKey, keyPath)
	}

	return privateKey, nil
}

func generatePrivateKey() (*ecdsa.PrivateKey, error) {
	pubkeyCurve := elliptic.P521()                     //see http://golang.org/pkg/crypto/elliptic/#P521
	return ecdsa.GenerateKey(pubkeyCurve, rand.Reader) // this generates a public & private key pair
}

func storePrivateKey(privateKey *ecdsa.PrivateKey, keyPath string) error {
	// write to key path
	ecder, errMarshalingKey := x509.MarshalECPrivateKey(privateKey)
	if errMarshalingKey != nil {
		return errMarshalingKey
	}
	// the private key, der encoded
	// TODO Check that reading the file has no issues with the order of the blocks, eg first params, then key blocks
	keypem, errSavingPEM := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if errSavingPEM != nil {
		return errSavingPEM
	}
	pem.Encode(keypem, &pem.Block{Type: "EC PRIVATE KEY", Bytes: ecder})
	// the elliptic curve parameters, der encoded
	// TODO These are hardcoded values for P256, we need to use the actual params
	// secp256r1, errAsnMarshal := asn1.Marshal(asn1.ObjectIdentifier{1, 2, 840, 10045, 3, 1, 7})
	// if errAsnMarshal != nil {
	// 	panic(errAsnMarshal)
	// }
	// pem.Encode(keypem, &pem.Block{Type: "EC PARAMETERS", Bytes: secp256r1})
	return nil
}

func loadPrivateKey(keyPath string) (*ecdsa.PrivateKey, error) {
	// read the file
	blockBytes, errOpeningFile := ioutil.ReadFile(keyPath)
	if errOpeningFile != nil {
		return nil, errOpeningFile
	}
	// decode the der blocks
	block, _ := pem.Decode(blockBytes)
	privateKey, errParsingPrivateKey := x509.ParseECPrivateKey(block.Bytes)
	if errParsingPrivateKey != nil {
		return nil, errParsingPrivateKey
	}
	return privateKey, nil
}

func Thumbprint(key *ecdsa.PrivateKey) string {
	h := sha1.New()
	ms := ethcrypto.FromECDSAPub(key.Public())
	h.Write(ms)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
