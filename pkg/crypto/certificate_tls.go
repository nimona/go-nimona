package crypto

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

// GenerateTLSCertificate for TLS serverset
func GenerateTLSCertificate(privateKey *PrivateKey) (*tls.Certificate, error) {
	k := privateKey.RawKey
	p := privateKey.PublicKey().RawKey
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(now.Unix()),
		Subject: pkix.Name{
			CommonName: privateKey.PublicKey().String(),
		},
		NotBefore: now,
		NotAfter:  now.AddDate(1, 0, 0),
		// TODO figure out the correct/best values for the following
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature |
			x509.KeyUsageCertSign,
	}
	cert, err := x509.CreateCertificate(
		rand.Reader,
		template,
		template,
		p,
		k,
	)
	if err != nil {
		return nil, err
	}

	outCert := &tls.Certificate{
		Certificate: [][]byte{
			cert,
		},
		PrivateKey: k,
	}
	return outCert, nil
}
