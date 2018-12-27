package hyperspace

import (
	"math"

	"github.com/james-bowman/sparse"
	"gonum.org/v1/gonum/mat"
)

// CosineSimilarity of two sparse vectors
func CosineSimilarity(a, b mat.Vector) float64 {
	dotProduct := sparse.Dot(a, b)
	norma := sparse.Norm(a, 2.0)
	normb := sparse.Norm(b, 2.0)

	if norma == 0 || normb == 0 {
		return math.NaN()
	}

	return (dotProduct / (norma * normb))
}
