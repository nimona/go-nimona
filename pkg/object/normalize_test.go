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

	s := stream.Created{
		Nonce:    "nonce",
		Identity: "foo",
		Policies: []*stream.Policy{
			&stream.Policy{
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
		"@type:s":           "nimona.io/stream.Created",
		"nonce:s":           "nonce",
		"createdDateTime:s": "",
		"@identity:s":       "foo",
		"policies:ao": []interface{}{
			map[string]interface{}{
				"@type:s":      "nimona.io/stream.Policy",
				"subjects:as":  []interface{}{"subject"},
				"resources:as": []interface{}{"*"},
				"action:s":     "allow",
			},
		},
		"@signature:o": map[string]interface{}{
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
