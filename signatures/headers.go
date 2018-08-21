package signatures

import (
	"github.com/nimona/go-nimona/blocks"
)

func init() {
	blocks.RegisterContentType("signature", Headers{})
}

// Headers for CWK
type Headers struct {
	Type      string `json:"typ"`
	Signature []byte `json:"sig"`
}
