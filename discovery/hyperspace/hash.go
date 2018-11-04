package hyperspace

import (
	"math"

	"github.com/spaolacci/murmur3"
)

const (
	murmurMin = uint32(0)
	murmurMax = uint32(math.MaxUint32)
	scaledMin = uint32(0)
	scaledMax = uint32(300)
)

// Hash finds the bucket of a given byte slice between scaleMin and scaleMax
// TODO(geoah) Rewrite, most likely wrong
func Hash(b []byte) int32 {
	v := murmur3.Sum32(b)
	p := float64(v-murmurMin) / float64(murmurMax-murmurMin)
	s := p*float64(scaledMax-scaledMin) + float64(scaledMin)
	return int32(s)
}
