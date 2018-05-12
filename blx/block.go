package blx

const (
	PayloadTypeTransferBlock string = "transfer-block"
	PayloadTypeRequestBlock         = "request-block"
)

type Block struct {
	Key    string            `json:"key"`
	Meta   map[string][]byte `json:"meta,omitempty"`
	Chunks []Chunk           `json:"chunks"`
}

type Chunk struct {
	BlockKey string `json:"block_key"`
	ChunkKey string `json:"chunck_key"`
	Data     []byte `json:"data"`
}

type payloadTransferBlock struct {
	Block *Block `json:"block,omitempty"`
}

type payloadTransferRequestBlock struct {
	Key string `json:"key,omitempty"`
}
