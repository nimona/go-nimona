package blx

const (
	PayloadTypeTransferBlock string = "transfer-block"
	PayloadTypeRequestBlock         = "request-block"
)

const (
	StatusOK = iota
	StatusNotFound
)

type Meta struct {
	Values map[string][]byte `json:"values,omitempty"`
}

type Block struct {
	Key  string `json:"key"`
	Meta Meta   `json:"meta"`
	Data []byte `json:"data"`
}

type payloadTransferBlock struct {
	Status int    `json:"status"`
	Nonce  string `json:"nonce"`
	Block  *Block `json:"block,omitempty"`
}

type payloadTransferRequestBlock struct {
	Nonce    string `json:"nonce"`
	Key      string `json:"key,omitempty"`
	response chan interface{}
}
