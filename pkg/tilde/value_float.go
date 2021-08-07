package tilde

import (
	"encoding/json"
	"fmt"
	"math"

	"nimona.io/pkg/errors"
)

func (v Float) Hint() Hint {
	return FloatHint
}

func (v Float) _isValue() {
}

func (v Float) Hash() Digest {
	// replacing ben's implementation with something less custom, based on:
	// * https://github.com/benlaurie/objecthash
	// * https://play.golang.org/p/3xraud43pi
	// examples of same results in other languages
	// * ruby: `[7.30363941192626953125].pack('G').unpack('B*').first`
	// * js: `http://weitz.de/ieee`
	//
	// NOTE(geoah): I have removed the inf and nan hashing for now,
	// we can revisit them once we better understand their usecases.
	switch {
	case math.IsInf(float64(v), 1):
		panic(errors.Error("float inf is not currently supported"))
	case math.IsInf(float64(v), -1):
		panic(errors.Error("float -inf is not currently supported"))
	case math.IsNaN(float64(v)):
		panic(errors.Error("float nan is not currently supported"))
	default:
		return hashFromBytes(
			[]byte(
				fmt.Sprintf(
					"%d",
					math.Float64bits(float64(v)),
				),
			),
		)
	}
}

func (v *Float) UnmarshalJSON(b []byte) error {
	var iv float64
	if err := json.Unmarshal(b, &iv); err != nil {
		return err
	}
	*v = Float(iv)
	return nil
}
