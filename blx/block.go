package blx

const (
	PayloadTypeTransferBlock string = "blx.transfer-block"
	PayloadTypeRequestBlock         = "blx.request-block"
)

const (
	StatusOK = iota
	StatusNotFound
)

// Meta is a struct that holds metadata for the specific block
type Meta struct {
	Values map[string][]byte `json:"values,omitempty"`
}

// Block is the base struct for data transmission and storage
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
	RequestingPeerID string `json:"req_peer_id"`
	Nonce            string `json:"nonce"`
	Key              string `json:"key,omitempty"`
	response         chan interface{}
}
