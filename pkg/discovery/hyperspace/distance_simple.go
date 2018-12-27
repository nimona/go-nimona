package hyperspace

import (
	"github.com/james-bowman/sparse"
)

// SimpleSimilarity of two sparse vectors
func SimpleSimilarity(a, b *sparse.Vector) float64 {
	commonCount := 0
	for i := 0; i < int(scaledMax); i++ {
		ia := a.AtVec(i)
		if ia == 0 {
			continue
		}
		if ia == b.AtVec(i) {
			commonCount++
		}
	}
	return float64(commonCount)
}
