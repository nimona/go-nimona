package fabric

// Basic imports
import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	suite "github.com/stretchr/testify/suite"
)

// ProtocolTLSTestSuite -
type ProtocolTLSTestSuite struct {
	suite.Suite
	ctx  context.Context
	cert tls.Certificate
}

func (suite *ProtocolTLSTestSuite) SetupTest() {
	suite.ctx = context.Background()
	cert, err := suite.createCert()
	suite.Assert().Nil(err)
	suite.cert = cert
}

func (suite *ProtocolTLSTestSuite) TestName() {
	tls := &SecProtocol{Config: tls.Config{
		Certificates:       []tls.Certificate{suite.cert},
		InsecureSkipVerify: true,
	}}

	name := tls.Name()
	suite.Assert().Equal("tls", name)
}

func (suite *ProtocolTLSTestSuite) createCert() (tls.Certificate, error) {
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
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, template,
		priv.Public(), priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	var outCert tls.Certificate
	outCert.Certificate = append(outCert.Certificate, cert)
	outCert.PrivateKey = priv

	return outCert, nil
}

func TestProtocolTLSTestSuite(t *testing.T) {
	suite.Run(t, new(ProtocolTLSTestSuite))
}
