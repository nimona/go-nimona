package crdt

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
)

const (
	EventTypeGraphSubscribe = "graph.subscribe"
	EventTypeGraphCreate    = "graph.create"
)

// Block -
type Block struct {
	Event     BlockEvent `json:"event"`
	Signature string     `json:"signature"`
	Hash      string     `json:"hash,omitempty"`
}

type BlockEvent struct {
	ACL struct {
		Read  []string `json:"read,omitempty"`
		Write []string `json:"write,omitempty"`
	} `json:"acl,omitempty"`
	Author  string   `json:"author"`
	Data    string   `json:"data"`
	Nonce   string   `json:"nonce"`
	Parents []string `json:"parents,omitempty"`
	Type    string   `json:"type"`
}

func HashBlock(block *Block) string {
	bs, _ := json.Marshal(block)
	return fmt.Sprintf("%x", sha1.Sum(bs))
}
