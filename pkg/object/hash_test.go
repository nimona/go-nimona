package object

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		hash  Hash
	}{
		{
			name:  "string",
			value: String("foo"),
			hash:  "oh1.Ff8LMmDZxqL7dn9SxiBusRxKDGQ5f7NzmnHJ7tEUFkgj",
		},
		{
			name:  "bytes",
			value: Bytes("foo"),
			hash:  "oh1.EuzQfFdCYuqmzZ5Htk56VzmYPYxdfbqPchqeps5d4q85",
		},
		{
			name:  "bool, false",
			value: Bool(false),
			hash:  "oh1.33SpDGfNXQXvojLuQbUeDbEjWeSTkWi1NUqS44kiUgSU",
		},
		{
			name:  "bool, true",
			value: Bool(true),
			hash:  "oh1.25Np3T8coGqqrbGwShY9xkY8VwYWP5xno3GwtrS7MZsa",
		},
		{
			name:  "int",
			value: Int(1234567890),
			hash:  "oh1.FEeWB9Uy6jTyx6eu2FUrhFEvi555ydaxwU3VUBjzi989",
		},
		{
			name:  "float",
			value: Float(12345.67890),
			hash:  "oh1.6YbTAHuuTfan9z8j6UVG9MYH8Y9oy9u6ZpiroFXoqXkT",
		},
		{
			name:  "float, +inf",
			value: Float(math.Inf(1)),
			hash:  "oh1.G69K5ZDfyi5yZEFhSseAp7tTBGWrZgc2jBYzqedxDyyg",
		},
		{
			name:  "float, -inf",
			value: Float(math.Inf(-1)),
			hash:  "oh1.2AwJmUEDPFy1cH2HMBQhjMjFpnq5SVQAteR4ZTxagCVC",
		},
		{
			name:  "float, nan",
			value: Float(math.NaN()),
			hash:  "oh1.7Hgbg5fn616DMPnim3sGDtZazXM1uBdKHXms4k7SKzwz",
		},
		{
			name: "array, string",
			value: List{}.
				Append(String("foo")).
				Append(String("bar")),
			hash: "oh1.HBLT1cYx761cbXHHxLtriS3afPUnoJ64USKRxBacers5",
		},
		{
			name: "array, int",
			value: List{}.
				Append(Int(123)).
				Append(Int(456)),
			hash: "oh1.H8rss8J56aDns3ZREEg9Mw1VwRkHtKq3yDqTuXB8aJV5",
		},
		{
			name: "map",
			value: Map{}.Set(
				"foo:s",
				String("bar"),
			),
			hash: "oh1.CgfoHRELcu1DwPjtGcXuVr1oFbAVxF8mRTWkTyJsE9gk",
		},
		{
			name: "map, nested",
			value: Map{}.Set(
				"data:o",
				Map{}.Set(
					"foo:s",
					String("bar"),
				),
			),
			hash: "oh1.EAKxMZySQigLYF9hZ3D4YjqrhWQ6q24NhvvbmUAsQSCt",
		},
		{
			name: "map, nested reference, should match previous",
			value: Map{}.Set(
				"data:r",
				Ref("oh1.CgfoHRELcu1DwPjtGcXuVr1oFbAVxF8mRTWkTyJsE9gk"),
			),
			hash: "oh1.EAKxMZySQigLYF9hZ3D4YjqrhWQ6q24NhvvbmUAsQSCt",
		},
	}
	for _, tt := range tests {
		got := tt.value.Hash()
		assert.Equal(t, tt.hash, got)
	}
}
