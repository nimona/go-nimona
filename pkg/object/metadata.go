package object

import (
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object/value"
)

// Metadata for object
type Metadata struct {
	Owner     crypto.PublicKey `nimona:"owner:s"`
	Datetime  string           `nimona:"datetime:s"`
	Parents   Parents          `nimona:"parents:m"`
	Policies  Policies         `nimona:"policies:am"`
	Stream    value.CID        `nimona:"stream:r"`
	Signature Signature        `nimona:"_signature:m"`
}
