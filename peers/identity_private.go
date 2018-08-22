package peers

import "github.com/nimona/go-nimona/blocks"

// PrivateIdentity holds the information for a private (local) identity
type PrivateIdentity struct {
	ID         string              `json:"id"`
	PrivateKey blocks.Key          `json:"private_key"`
	Peers      *PeerInfoCollection `json:"-"`
}
