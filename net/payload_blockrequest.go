package net

const (
	PayloadTypeRequestBlock = "blx.request-block"
)

func init() {
	RegisterContentType(PayloadTypeRequestBlock, PayloadRequestBlock{})
}

// PayloadRequestBlock payload for PayloadTypeRequestBlock
type PayloadRequestBlock struct {
	RequestID string `json:"requestID"`
	ID        string `json:"id"`
	response  chan *Block
}
