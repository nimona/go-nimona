package nimona

import "github.com/mr-tron/base58/base58"

type (
	PublicKey  []byte
	PrivateKey []byte
)

func (k PublicKey) String() string {
	return base58.Encode(k)
}
