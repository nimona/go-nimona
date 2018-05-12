package blx

const (
	PayloadTypeTransferBlock string = "transfer-block"
	PayloadTypeRequestBlock         = "request-block"
)

type Block struct {
	Key  string            `json:"key"`
	Meta map[string][]byte `json:"meta,omitempty"`
	Data []byte            `json:"data"`
}

type payloadTransferBlock struct {
	Block *Block `json:"block,omitempty"`
}

type payloadTransferRequestBlock struct {
	Key string `json:"key,omitempty"`
}
