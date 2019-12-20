package object_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"nimona.io/internal/fixtures"
	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
)

func Test_Normalize(t *testing.T) {
	kx := []byte{111, 39, 65, 188, 215, 49, 84, 43, 192, 187}

	s := fixtures.TestStream{
		Nonce:    "nonce",
		Identity: "foo",
		Policies: []*fixtures.TestPolicy{
			&fixtures.TestPolicy{
				Subjects:  []string{"subject"},
				Resources: []string{"*"},
				Action:    "allow",
			},
		},
		Signature: &crypto.Signature{
			Signer: "foo",
			Alg:    "alg",
			X:      kx,
		},
	}

	em := map[string]interface{}{
		"$schema:o": map[string]interface{}{
			"@type:s": string("nimona.io/schema.Object"),
			"properties:ao": []interface{}{
				map[string]interface{}{
					"@type:s":      string("nimona.io/schema.Property"),
					"hint:s":       string("s"),
					"isOptional:b": bool(false),
					"isRepeated:b": bool(false),
					"name:s":       string("nonce"),
					"type:s":       string("string"),
				},
				map[string]interface{}{
					"@type:s":      string("nimona.io/schema.Property"),
					"hint:s":       string("s"),
					"isOptional:b": bool(false),
					"isRepeated:b": bool(false),
					"name:s":       string("createdDateTime"),
					"type:s":       string("string"),
				},
				map[string]interface{}{
					"@type:s":      string("nimona.io/schema.Property"),
					"hint:s":       string("o"),
					"isOptional:b": bool(false),
					"isRepeated:b": bool(true),
					"name:s":       string("policies"),
					"type:s":       string("TestPolicy"),
				},
				map[string]interface{}{
					"@type:s":      string("nimona.io/schema.Property"),
					"hint:s":       string("o"),
					"isOptional:b": bool(false),
					"isRepeated:b": bool(false),
					"name:s":       string("@signature"),
					"type:s":       string("nimona.io/crypto.Signature"),
				},
				map[string]interface{}{
					"@type:s":      string("nimona.io/schema.Property"),
					"hint:s":       string("s"),
					"isOptional:b": bool(false),
					"isRepeated:b": bool(false),
					"name:s":       string("@identity"),
					"type:s":       string("nimona.io/crypto.PublicKey"),
				},
			},
		},
		"@type:s":     "nimona.io/fixtures.TestStream",
		"nonce:s":     "nonce",
		"@identity:s": "foo",
		"policies:ao": []interface{}{
			map[string]interface{}{
				"@type:s":      "nimona.io/fixtures.TestPolicy",
				"subjects:as":  []interface{}{"subject"},
				"resources:as": []interface{}{"*"},
				"action:s":     "allow",
				"$schema:o": map[string]interface{}{
					"@type:s": string("nimona.io/schema.Object"),
					"properties:ao": []interface{}{
						map[string]interface{}{
							"@type:s":      string("nimona.io/schema.Property"),
							"hint:s":       string("s"),
							"isOptional:b": bool(false),
							"isRepeated:b": bool(true),
							"name:s":       string("subjects"),
							"type:s":       string("string"),
						},
						map[string]interface{}{
							"@type:s":      string("nimona.io/schema.Property"),
							"hint:s":       string("s"),
							"isOptional:b": bool(false),
							"isRepeated:b": bool(true),
							"name:s":       string("resources"),
							"type:s":       string("string"),
						},
						map[string]interface{}{
							"@type:s":      string("nimona.io/schema.Property"),
							"hint:s":       string("s"),
							"isOptional:b": bool(false),
							"isRepeated:b": bool(true),
							"name:s":       string("conditions"),
							"type:s":       string("string"),
						},
						map[string]interface{}{
							"@type:s":      string("nimona.io/schema.Property"),
							"hint:s":       string("s"),
							"isOptional:b": bool(false),
							"isRepeated:b": bool(false),
							"name:s":       string("action"),
							"type:s":       string("string"),
						},
					},
				},
			},
		},
		"@signature:o": map[string]interface{}{
			"$schema:o": map[string]interface{}{
				"@type:s": string("nimona.io/schema.Object"),
				"properties:ao": []interface{}{
					map[string]interface{}{
						"@type:s":      string("nimona.io/schema.Property"),
						"hint:s":       string("s"),
						"isOptional:b": bool(false),
						"isRepeated:b": bool(false),
						"name:s":       string("signer"),
						"type:s":       string("nimona.io/crypto.PublicKey"),
					},
					map[string]interface{}{
						"@type:s":      string("nimona.io/schema.Property"),
						"hint:s":       string("s"),
						"isOptional:b": bool(false),
						"isRepeated:b": bool(false),
						"name:s":       string("alg"),
						"type:s":       string("string"),
					},
					map[string]interface{}{
						"@type:s":      string("nimona.io/schema.Property"),
						"hint:s":       string("d"),
						"isOptional:b": bool(false),
						"isRepeated:b": bool(false),
						"name:s":       string("x"),
						"type:s":       string("data"),
					},
				},
			},
			"@type:s":  "nimona.io/crypto.Signature",
			"signer:s": "foo",
			"alg:s":    "alg",
			"x:d":      kx,
		},
	}

	b, err := json.MarshalIndent(s.ToObject().ToMap(), "", "  ")
	require.NoError(t, err)

	m := map[string]interface{}{}
	require.NoError(t, json.Unmarshal(b, &m))

	nm, err := object.Normalize(m)
	require.NoError(t, err)

	if !cmp.Equal(nm, em) {
		t.Errorf("Normalize() result doesn't match expectd " + cmp.Diff(nm, em))
	}
}
