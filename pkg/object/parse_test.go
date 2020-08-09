package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseType(t *testing.T) {
	tests := []struct {
		objectType string
		want       ParsedType
	}{{
		objectType: "stream:nimona.io/conversation",
		want: ParsedType{
			PrimaryType: "stream",
			Namespace:   "nimona.io",
			Object:      "conversation",
		},
	}, {
		objectType: "event:nimona.io/conversation.AddMessage",
		want: ParsedType{
			PrimaryType: "event",
			Namespace:   "nimona.io",
			Object:      "conversation.AddMessage",
		},
	}, {
		objectType: "nimona.io/PublicKey",
		want: ParsedType{
			PrimaryType: "",
			Namespace:   "nimona.io",
			Object:      "PublicKey",
		},
	}, {
		objectType: "nimona.io/crypto.PublicKey",
		want: ParsedType{
			PrimaryType: "",
			Namespace:   "nimona.io",
			Object:      "crypto.PublicKey",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.objectType, func(t *testing.T) {
			got := ParseType(tt.objectType)
			assert.Equal(t, tt.want, got)
		})
	}
}
