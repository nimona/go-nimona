package crypto

import (
	"math/big"
)

func bigIntFromBytes(b []byte) *big.Int {
	i := &big.Int{}
	return i.SetBytes(b)
}
