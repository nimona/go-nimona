package blx

import (
	"github.com/nimona/go-nimona/net"
)

const (
	// PayloadTypeTransferBlock type for PayloadTransferBlock
	PayloadTypeTransferBlock string = "blx.transfer-block"
	// PayloadTypeRequestBlock type for PayloadRequestBlock
	PayloadTypeRequestBlock = "blx.request-block"
)

func init() {
	net.RegisterContentType(PayloadTypeTransferBlock, PayloadTransferBlock{})
	net.RegisterContentType(PayloadTypeRequestBlock, PayloadRequestBlock{})
}

const (
	// StatusOK when block exists and is healthy
	StatusOK = "ok"
	// StatusNotFound when a block does not exist
	StatusNotFound = "not-found"
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

// PayloadTransferBlock payload for PayloadTypeTransferBlock
type PayloadTransferBlock struct {
	Status string `json:"status"`
	Nonce  string `json:"nonce"`
	Block  *Block `json:"block,omitempty"`
}

// PayloadRequestBlock payload for PayloadTypeRequestBlock
type PayloadRequestBlock struct {
	RequestingPeerID string `json:"req_peer_id"`
	Nonce            string `json:"nonce"`
	Key              string `json:"key,omitempty"`
	response         chan interface{}
}
