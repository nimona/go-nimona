package immutable

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		hash  string
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
			hash: "oh1.2uMKwin1K763GKkFuKjToDPNKRFxDEJ2JMoxACerEGtT",
		},
		{
			name: "array, int",
			value: List{}.
				Append(Int(123)).
				Append(Int(456)),
			hash: "oh1.FSV8pEhavs8GzYW7iH9WxC69SxnSruGmqNKny1mttitj",
		},
		{
			name: "map",
			value: Map{}.Set(
				"foo:s",
				String("bar"),
			),
			hash: "oh1.A2Q9xKrKqjvr1KLr32AcMJBGiMZ5UpKpuHojjhafLkEj",
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
			hash: "oh1.DmjgHp5UYzXV9EUjcbrsdXDV2agUTQH5YsosULNZRwtM",
		},
	}
	for _, tt := range tests {
		got := tt.value.Hash()
		assert.Equal(t, tt.hash, got)
	}
}
