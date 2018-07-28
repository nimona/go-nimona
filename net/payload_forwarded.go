package net

const (
	TypeForwarded = "nimona.forwarded"
)

func init() {
	RegisterContentType(TypeForwarded, PayloadForwarded{})
}

// PayloadForwarded is the payload for proxied blocks
type PayloadForwarded struct {
	RecipientID string `json:"recipientID"`
	Block       *Block `json:"block"`
}
