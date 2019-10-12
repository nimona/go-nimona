package object_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"nimona.io/pkg/crypto"
	"nimona.io/pkg/object"
	"nimona.io/pkg/stream"
)

func Test_Normalize(t *testing.T) {
	kx := []byte{111, 39, 65, 188, 215, 49, 84, 43, 192, 187}
	ky := []byte{235, 143, 194, 90, 199, 139, 81, 230, 181, 145}
	sr := []byte{140, 7, 104, 131, 111, 142, 147, 105, 153, 234}
	ss := []byte{98, 144, 226, 154, 141, 254, 84, 218, 191, 16}

	s := stream.Created{
		Nonce: "nonce",
		Authors: []*stream.Author{
			&stream.Author{
				PublicKey: &crypto.PublicKey{
					KeyType:   "kty",
					Algorithm: "alg",
					Curve:     "crv",
					X:         kx,
					Y:         ky,
					Signature: &crypto.Signature{
						Algorithm: "alg",
						R:         sr,
						S:         ss,
					},
				},
			},
		},
		Policies: []*stream.Policy{
			&stream.Policy{
				Subjects:  []string{"subject"},
				Resources: []string{"*"},
				Action:    "allow",
			},
		},
		Signature: &crypto.Signature{
			PublicKey: &crypto.PublicKey{
				KeyType:   "kty",
				Algorithm: "alg",
				Curve:     "crv",
				X:         kx,
				Y:         ky,
			},
			Algorithm: "alg",
			R:         sr,
			S:         ss,
		},
	}

	em := map[string]interface{}{
		"@ctx:s":            "nimona.io/stream.Created",
		"nonce:s":           "nonce",
		"createdDateTime:s": "",
		"@authors:ao": []interface{}{
			map[string]interface{}{
				"@ctx:s": "nimona.io/stream.Author",
				"publicKey:o": map[string]interface{}{
					"@ctx:s":      "nimona.io/crypto.PublicKey",
					"keyType:s":   "kty",
					"algorithm:s": "alg",
					"curve:s":     "crv",
					"x:d":         kx,
					"y:d":         ky,
					"@signature:o": map[string]interface{}{
						"@ctx:s":      "nimona.io/crypto.Signature",
						"algorithm:s": "alg",
						"r:d":         sr,
						"s:d":         ss,
					},
				},
			},
		},
		"@policies:ao": []interface{}{
			map[string]interface{}{
				"@ctx:s":       "nimona.io/stream.Policy",
				"subjects:as":  []interface{}{"subject"},
				"resources:as": []interface{}{"*"},
				"action:s":     "allow",
			},
		},
		"@signature:o": map[string]interface{}{
			"@ctx:s": "nimona.io/crypto.Signature",
			"publicKey:o": map[string]interface{}{
				"@ctx:s":      "nimona.io/crypto.PublicKey",
				"keyType:s":   "kty",
				"algorithm:s": "alg",
				"curve:s":     "crv",
				"x:d":         kx,
				"y:d":         ky,
			},
			"algorithm:s": "alg",
			"r:d":         sr,
			"s:d":         ss,
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
