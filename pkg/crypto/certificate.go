package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"

	"github.com/pkg/errors"
)

// GenerateCertificate for TLS serverset
func GenerateCertificate(key *PrivateKey) (*tls.Certificate, error) {

	pk, ok := key.Key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("only ecdsa private keys are supported")
	}

	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(now.Unix()),
		Subject: pkix.Name{
			CommonName:         "quickserve.example.com",
			Country:            []string{"USA"},
			Organization:       []string{"example.com"},
			OrganizationalUnit: []string{"quickserve"},
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 0, 1), // Valid for one day
		SubjectKeyId:          []byte{113, 117, 105, 99, 107, 115, 101, 114, 118, 101},
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	// priv, err := rsa.GenerateKey(rand.Reader, 2048)
	// if err != nil {
	// 	return nil, err
	// }

	cert, err := x509.CreateCertificate(rand.Reader, template, template,
		pk.Public(), pk)
	if err != nil {
		return nil, err
	}

	var outCert tls.Certificate
	outCert.Certificate = append(outCert.Certificate, cert)
	outCert.PrivateKey = pk

	return &outCert, nil
}
