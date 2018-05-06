package blx

const (
	PayloadTypeTransferBlock string = "transfer-block"
	PayloadTypeRequestBlock         = "request-block"
)

type Block struct {
	Key  string            `json:"key,omitempty"`
	Data []byte            `json:"data,omitempty"`
	Meta map[string][]byte `json:"meta,omitempty"`
}

type payloadTransferBlock struct {
	Block *Block `json:"block,omitempty"`
}

type payloadTransferRequestBlock struct {
	Key string `json:"key,omitempty"`
}
