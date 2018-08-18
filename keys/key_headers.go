package keys

import (
	"github.com/nimona/go-nimona/blocks"
)

func init() {
	blocks.RegisterContentType("key", Headers{})
}

// Headers for CWK
type Headers struct {
	Algorithm              string `json:"alg,omitempty"`
	KeyID                  string `json:"kid,omitempty"`
	KeyType                string `json:"kty,omitempty"`
	KeyUsage               string `json:"use,omitempty"`
	KeyOps                 string `json:"key_ops,omitempty"`
	X509CertChain          string `json:"x5c,omitempty"`
	X509CertThumbprint     string `json:"x5t,omitempty"`
	X509CertThumbprintS256 string `json:"x5t#S256,omitempty"`
	X509URL                string `json:"x5u,omitempty"`
	Curve                  string `json:"crv,omitempty"`
	X                      []byte `json:"x,omitempty"`
	Y                      []byte `json:"y,omitempty"`
	D                      []byte `json:"d,omitempty"`
}
