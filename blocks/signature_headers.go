package blocks

func init() {
	RegisterContentType("signature", SignatureHeaders{})
}

// SignatureHeaders for CWK
type SignatureHeaders struct {
	Type      string `json:"typ"`
	Signature []byte `json:"sig"`
}
