package hyperspace

import (
	"fmt"
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

// HashChunked spilts input in chunks and hashes them individually
// TODO(geoah) Rewrite, most likely wrong as well
func HashChunked(prefix string, o []byte) []int {
	i := []int{}
	var b int32
	for j, c := range chunk(o, 4) {
		b = Hash([]byte(fmt.Sprintf("%s_%d_%s", prefix, j, string(c))))
		i = append(i, int(b))
	}
	return i
}
